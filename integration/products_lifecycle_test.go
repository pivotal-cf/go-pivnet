package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Products Lifecycle", func() {
	Describe("finding a product by slug", func() {
		It("returns the corresponding product", func() {
			products, err := client.Releases.List(testProductSlug)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(products)).NotTo(Equal(0))
		})
	})
})
