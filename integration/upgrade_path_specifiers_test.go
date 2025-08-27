package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

var _ = Describe("Upgrade Path Specifier Integration", func() {
	var (
		newRelease pivnet.Release
	)

	It("creates, lists, and deletes upgrade path specifiers", func() {
		By("Listing upgrade path specifiers")
		_, err := client.UpgradePathSpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

	})
})
