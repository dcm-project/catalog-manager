BINARY_NAME := catalog-manager

build:
	go build -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

run:
	go run ./cmd/$(BINARY_NAME)

clean:
	rm -rf bin/

fmt:
	gofmt -s -w .

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

generate-types:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/types.gen.cfg \
		-o api/v1alpha1/types.gen.go \
		api/v1alpha1/openapi.yaml

generate-spec:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/spec.gen.cfg \
		-o api/v1alpha1/spec.gen.go \
		api/v1alpha1/openapi.yaml

generate-server:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=internal/api/server/server.gen.cfg \
		-o internal/api/server/server.gen.go \
		api/v1alpha1/openapi.yaml

generate-client:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=pkg/client/client.gen.cfg \
		-o pkg/client/client.gen.go \
		api/v1alpha1/openapi.yaml

generate-api: generate-types generate-spec generate-server generate-client

check-generate-api: generate-api
	git diff --exit-code api/ internal/api/server/ pkg/client/ || \
		(echo "Generated files out of sync. Run 'make generate-api'." && exit 1)

# Check AEP compliance
check-aep:
	spectral lint --fail-severity=warn ./api/v1alpha1/openapi.yaml


# Generate Go types for service specifications (VM, Container, Database, Cluster)
generate-service-types:
	@echo "Generating common types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/types.gen.cfg \
		-o api/v1alpha1/servicetypes/types.gen.go \
		api/v1alpha1/servicetypes/common.yaml
	@echo "Generating VM types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/vm/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/vm/types.gen.go \
		api/v1alpha1/servicetypes/vm/spec.yaml
	@echo "Generating Container types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/container/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/container/types.gen.go \
		api/v1alpha1/servicetypes/container/spec.yaml
	@echo "Generating Database types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/database/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/database/types.gen.go \
		api/v1alpha1/servicetypes/database/spec.yaml
	@echo "Generating Cluster types..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		--config=api/v1alpha1/servicetypes/cluster/spec.gen.cfg \
		--import-mapping=../common.yaml:github.com/dcm-project/catalog-manager/api/v1alpha1/servicetypes \
		-o api/v1alpha1/servicetypes/cluster/types.gen.go \
		api/v1alpha1/servicetypes/cluster/spec.yaml
	@echo "Service types generation complete!"

.PHONY: build run clean fmt vet test tidy generate-types generate-spec generate-server generate-client generate-api check-generate-api check-aep generate-service-types