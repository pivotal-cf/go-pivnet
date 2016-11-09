package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet"
)

var _ = Describe("Products Lifecycle", func() {
	Describe("finding a product by slug", func() {
		It("returns the corresponding product", func() {
			product, err := client.Products.Get(testProductSlug)
			Expect(err).NotTo(HaveOccurred())

			Expect(product).To(Equal(pivnet.Product{
				ID:   90,
				Slug: testProductSlug,
				Name: "Pivnet Resource Test",
			}))
		})
	})
})
