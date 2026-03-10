package service_test

import (
	"context"
	"log/slog"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dcm-project/catalog-manager/api/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/service"
	"github.com/dcm-project/catalog-manager/internal/store"
	"github.com/dcm-project/catalog-manager/internal/store/model"
)

var _ = Describe("SpecBuilder (via CatalogItemInstance Create)", func() {
	var (
		ctx context.Context
		db  *gorm.DB
		str store.Store
		svc service.Service
	)

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
			Logger: logger.Discard,
		})
		Expect(err).ToNot(HaveOccurred())
		err = db.Exec("PRAGMA foreign_keys = ON").Error
		Expect(err).ToNot(HaveOccurred())
		err = db.AutoMigrate(&model.ServiceType{}, &model.CatalogItem{}, &model.CatalogItemInstance{})
		Expect(err).ToNot(HaveOccurred())
		str = store.NewStore(db, slog.Default())
		svc = service.NewService(str, nil, slog.Default())

		// Seed ServiceType with a rich spec
		ensureServiceTypeWithSpec(ctx, str, "vm-spec-builder", "vm-sb", map[string]any{
			"vcpu":   map[string]any{"count": float64(1)},
			"memory": map[string]any{"size_gb": float64(2)},
			"disk":   map[string]any{"size_gb": float64(50)},
		})
	})

	AfterEach(func() {
		if str != nil {
			Expect(str.Close()).To(Succeed())
		}
	})

	Describe("full chain resolution", func() {
		It("should resolve ServiceType → CatalogItem → Instance", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-chain", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(4), Editable: true},
				{Path: "spec.memory.size_gb", Default: float64(8), Editable: true},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "Chain Test",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-chain",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.vcpu.count", Value: float64(16)},
					},
				},
			}

			result, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})

	Describe("validation errors", func() {
		It("should reject user_value path that doesn't match any CatalogItem field", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-bad-path", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(2), Editable: true},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "Bad Path",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-bad-path",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.network.bandwidth", Value: float64(100)},
					},
				},
			}

			_, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("user value path not found"))
		})

		It("should reject user_value for non-editable field", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-not-editable", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.disk.size_gb", Default: float64(50), Editable: false},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "Not Editable",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-not-editable",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.disk.size_gb", Value: float64(100)},
					},
				},
			}

			_, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not editable"))
		})

		It("should reject user_value that fails validation_schema", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-schema-fail", "vm-sb", []model.FieldConfiguration{
				{
					Path:     "spec.vcpu.count",
					Default:  float64(2),
					Editable: true,
					ValidationSchema: map[string]any{
						"type":    "integer",
						"minimum": float64(1),
						"maximum": float64(16),
					},
				},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "Schema Fail",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-schema-fail",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.vcpu.count", Value: float64(32)},
					},
				},
			}

			_, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("validation failed"))
		})

		It("should accept user_value that passes validation_schema", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-schema-pass", "vm-sb", []model.FieldConfiguration{
				{
					Path:     "spec.vcpu.count",
					Default:  float64(2),
					Editable: true,
					ValidationSchema: map[string]any{
						"type":    "integer",
						"minimum": float64(1),
						"maximum": float64(16),
					},
				},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "Schema Pass",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-schema-pass",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.vcpu.count", Value: float64(8)},
					},
				},
			}

			result, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should reject user_value that violates depends_on constraint", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-depends-fail", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(2), Editable: true},
				{
					Path:     "spec.memory.size_gb",
					Default:  float64(4),
					Editable: true,
					DependsOn: &model.DependsOn{
						Path: "spec.vcpu.count",
						AllowedValues: map[string][]any{
							"2": {float64(4), float64(8)},
							"4": {float64(8), float64(16)},
						},
					},
				},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "DependsOn Fail",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-depends-fail",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.memory.size_gb", Value: float64(32)},
					},
				},
			}

			_, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("depends_on"))
		})

		It("should accept user_value that satisfies depends_on constraint", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-depends-pass", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(2), Editable: true},
				{
					Path:     "spec.memory.size_gb",
					Default:  float64(4),
					Editable: true,
					DependsOn: &model.DependsOn{
						Path: "spec.vcpu.count",
						AllowedValues: map[string][]any{
							"2": {float64(4), float64(8)},
							"4": {float64(8), float64(16)},
						},
					},
				},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "DependsOn Pass",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-depends-pass",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.memory.size_gb", Value: float64(8)},
					},
				},
			}

			result, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should validate depends_on against updated source value from user_values", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-depends-updated", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(2), Editable: true},
				{
					Path:     "spec.memory.size_gb",
					Default:  float64(4),
					Editable: true,
					DependsOn: &model.DependsOn{
						Path: "spec.vcpu.count",
						AllowedValues: map[string][]any{
							"2": {float64(4), float64(8)},
							"4": {float64(8), float64(16)},
						},
					},
				},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "DependsOn Updated Source",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-depends-updated",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.vcpu.count", Value: float64(4)},
						{Path: "spec.memory.size_gb", Value: float64(16)},
					},
				},
			}

			result, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should validate depends_on with source field listed after dependent in user_values", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-depends-order", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(2), Editable: true},
				{
					Path:     "spec.memory.size_gb",
					Default:  float64(4),
					Editable: true,
					DependsOn: &model.DependsOn{
						Path: "spec.vcpu.count",
						AllowedValues: map[string][]any{
							"4": {float64(8), float64(16)},
						},
					},
				},
			})

			// memory depends on vcpu, but memory is listed first — must still validate against vcpu=4
			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "DependsOn Reverse Order",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-depends-order",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.memory.size_gb", Value: float64(16)},
						{Path: "spec.vcpu.count", Value: float64(4)},
					},
				},
			}

			result, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should reject depends_on when source value has no allowed_values entry", func() {
			ensureCatalogItemWithFields(ctx, str, "ci-depends-no-key", "vm-sb", []model.FieldConfiguration{
				{Path: "spec.vcpu.count", Default: float64(2), Editable: true},
				{
					Path:     "spec.memory.size_gb",
					Default:  float64(4),
					Editable: true,
					DependsOn: &model.DependsOn{
						Path: "spec.vcpu.count",
						AllowedValues: map[string][]any{
							"2": {float64(4), float64(8)},
						},
					},
				},
			})

			req := &service.CreateCatalogItemInstanceRequest{
				ApiVersion:  "v1alpha1",
				DisplayName: "DependsOn No Key",
				Spec: v1alpha1.CatalogItemInstanceSpec{
					CatalogItemId: "ci-depends-no-key",
					UserValues: []v1alpha1.UserValue{
						{Path: "spec.vcpu.count", Value: float64(8)},
						{Path: "spec.memory.size_gb", Value: float64(4)},
					},
				},
			}

			_, err := svc.CatalogItemInstance().Create(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no allowed values defined"))
		})
	})
})
