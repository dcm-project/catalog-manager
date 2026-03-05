package service

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nested Map Utilities", func() {
	Describe("stripSpecPrefix", func() {
		It("strips the spec. prefix", func() {
			Expect(stripSpecPrefix("spec.vcpu.count")).To(Equal("vcpu.count"))
			Expect(stripSpecPrefix("spec.memory.size_gb")).To(Equal("memory.size_gb"))
		})

		It("leaves paths without spec. prefix unchanged", func() {
			Expect(stripSpecPrefix("vcpu.count")).To(Equal("vcpu.count"))
			Expect(stripSpecPrefix("metadata.labels.tier")).To(Equal("metadata.labels.tier"))
		})

		It("returns empty string for bare spec. prefix", func() {
			Expect(stripSpecPrefix("spec.")).To(Equal(""))
		})
	})

	Describe("setNestedValue", func() {
		var m map[string]any

		BeforeEach(func() {
			m = make(map[string]any)
		})

		It("sets value at top level", func() {
			err := setNestedValue(m, "spec.vcpu", 4)
			Expect(err).ToNot(HaveOccurred())
			Expect(m["vcpu"]).To(Equal(4))
		})

		It("sets value at nested path", func() {
			err := setNestedValue(m, "spec.vcpu.count", 4)
			Expect(err).ToNot(HaveOccurred())
			vcpu, ok := m["vcpu"].(map[string]any)
			Expect(ok).To(BeTrue())
			Expect(vcpu["count"]).To(Equal(4))
		})

		It("sets value into existing nested structure", func() {
			m["vcpu"] = map[string]any{"count": 2}
			err := setNestedValue(m, "spec.vcpu.count", 8)
			Expect(err).ToNot(HaveOccurred())
			vcpu := m["vcpu"].(map[string]any)
			Expect(vcpu["count"]).To(Equal(8))
		})

		It("creates deeply nested path", func() {
			err := setNestedValue(m, "spec.a.b.c.d", "deep")
			Expect(err).ToNot(HaveOccurred())
			a := m["a"].(map[string]any)
			b := a["b"].(map[string]any)
			c := b["c"].(map[string]any)
			Expect(c["d"]).To(Equal("deep"))
		})

		It("returns error when intermediate is not a map", func() {
			m["vcpu"] = "not-a-map"
			err := setNestedValue(m, "spec.vcpu.count", 4)
			Expect(err).To(HaveOccurred())
		})

		It("works without spec prefix", func() {
			err := setNestedValue(m, "vcpu.count", 4)
			Expect(err).ToNot(HaveOccurred())
			vcpu := m["vcpu"].(map[string]any)
			Expect(vcpu["count"]).To(Equal(4))
		})

		Context("with different value types", func() {
			It("sets a boolean value", func() {
				err := setNestedValue(m, "spec.gpu.enabled", true)
				Expect(err).ToNot(HaveOccurred())
				gpu := m["gpu"].(map[string]any)
				Expect(gpu["enabled"]).To(Equal(true))
			})

			It("sets a false boolean value", func() {
				err := setNestedValue(m, "spec.gpu.enabled", false)
				Expect(err).ToNot(HaveOccurred())
				gpu := m["gpu"].(map[string]any)
				Expect(gpu["enabled"]).To(Equal(false))
			})

			It("sets a float64 value", func() {
				err := setNestedValue(m, "spec.memory.size_gb", 16.5)
				Expect(err).ToNot(HaveOccurred())
				memory := m["memory"].(map[string]any)
				Expect(memory["size_gb"]).To(Equal(16.5))
			})

			It("sets a nil value", func() {
				err := setNestedValue(m, "spec.metadata.description", nil)
				Expect(err).ToNot(HaveOccurred())
				metadata := m["metadata"].(map[string]any)
				Expect(metadata["description"]).To(BeNil())
			})

			It("sets a slice value", func() {
				tags := []string{"prod", "gpu", "high-mem"}
				err := setNestedValue(m, "spec.metadata.tags", tags)
				Expect(err).ToNot(HaveOccurred())
				metadata := m["metadata"].(map[string]any)
				Expect(metadata["tags"]).To(Equal(tags))
			})

			It("sets a nested map value", func() {
				labels := map[string]any{"tier": "premium", "region": "us-east"}
				err := setNestedValue(m, "spec.metadata.labels", labels)
				Expect(err).ToNot(HaveOccurred())
				metadata := m["metadata"].(map[string]any)
				Expect(metadata["labels"]).To(Equal(labels))
			})
		})

		Context("with multiple fields set on the same map", func() {
			It("sets sibling fields without overwriting each other", func() {
				Expect(setNestedValue(m, "spec.vcpu.count", 4)).To(Succeed())
				Expect(setNestedValue(m, "spec.vcpu.frequency_ghz", 3.2)).To(Succeed())
				Expect(setNestedValue(m, "spec.vcpu.hyperthreading", true)).To(Succeed())

				vcpu := m["vcpu"].(map[string]any)
				Expect(vcpu).To(HaveLen(3))
				Expect(vcpu["count"]).To(Equal(4))
				Expect(vcpu["frequency_ghz"]).To(Equal(3.2))
				Expect(vcpu["hyperthreading"]).To(Equal(true))
			})

			It("sets fields across different subtrees", func() {
				Expect(setNestedValue(m, "spec.vcpu.count", 8)).To(Succeed())
				Expect(setNestedValue(m, "spec.memory.size_gb", 32.0)).To(Succeed())
				Expect(setNestedValue(m, "spec.gpu.enabled", false)).To(Succeed())

				Expect(m["vcpu"].(map[string]any)["count"]).To(Equal(8))
				Expect(m["memory"].(map[string]any)["size_gb"]).To(Equal(32.0))
				Expect(m["gpu"].(map[string]any)["enabled"]).To(Equal(false))
			})
		})

		Context("with error cases", func() {
			It("returns error for empty path and sets nothing", func() {
				err := setNestedValue(m, "", "should-not-be-set")
				Expect(err).To(HaveOccurred())
				Expect(m).To(BeEmpty())
			})

			It("returns error when an integer intermediate blocks traversal", func() {
				m["vcpu"] = 42
				err := setNestedValue(m, "spec.vcpu.count", 4)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("vcpu"))
			})

			It("returns error when a boolean intermediate blocks traversal", func() {
				m["gpu"] = true
				err := setNestedValue(m, "spec.gpu.model.name", "A100")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("gpu"))
			})

			It("returns error when a deeply nested intermediate is not a map", func() {
				m["a"] = map[string]any{"b": "leaf"}
				err := setNestedValue(m, "spec.a.b.c.d", "value")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("a.b"))
			})
		})

		It("overwrites an existing value with a different type", func() {
			Expect(setNestedValue(m, "spec.vcpu.count", 4)).To(Succeed())
			Expect(setNestedValue(m, "spec.vcpu.count", "four")).To(Succeed())
			vcpu := m["vcpu"].(map[string]any)
			Expect(vcpu["count"]).To(Equal("four"))
		})

		It("handles a single-segment path without spec prefix", func() {
			err := setNestedValue(m, "name", "my-instance")
			Expect(err).ToNot(HaveOccurred())
			Expect(m["name"]).To(Equal("my-instance"))
		})
	})

})
