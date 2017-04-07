package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf/go-pivnet"
)

const (
	upgradePathSpecifier = "~>1.2.3"
)

var _ = Describe("Upgrade Path Specifier Integration", func() {
	var (
		newRelease pivnet.Release
	)

	BeforeEach(func() {
		var err error
		newRelease, err = client.Releases.Create(pivnet.CreateReleaseConfig{
			ProductSlug: testProductSlug,
			Version:     "some-test-version",
			ReleaseType: "Beta Release",
			EULASlug:    "pivotal_beta_eula",
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := client.Releases.Delete(testProductSlug, newRelease)
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates, lists, and deletes upgrade path specifiers", func() {
		By("Listing upgrade path specifiers")
		upgradePathSpecifiers, err := client.UpgradePathSpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

		By("Creating new dependency specifier")
		upgradePathSpecifier, err := client.UpgradePathSpecifiers.Create(
			testProductSlug,
			newRelease.ID,
			upgradePathSpecifier,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(upgradePathSpecifiers).ShouldNot(ContainElement(upgradePathSpecifier))

		By("Re-listing upgrade path specifiers")
		updatedUpgradePathSpecifiers, err := client.UpgradePathSpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

		Expect(updatedUpgradePathSpecifiers).Should(ContainElement(upgradePathSpecifier))

		By("Getting individual dependency specifier")
		individualUpgradePathSpecifier, err := client.UpgradePathSpecifiers.Get(
			testProductSlug,
			newRelease.ID,
			upgradePathSpecifier.ID,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(individualUpgradePathSpecifier).To(Equal(upgradePathSpecifier))

		By("Deleting upgrade path specifier")
		err = client.UpgradePathSpecifiers.Delete(
			testProductSlug,
			newRelease.ID,
			upgradePathSpecifier.ID,
		)
		Expect(err).NotTo(HaveOccurred())

		newlyUpdatedUpgradePathSpecifiers, err := client.UpgradePathSpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

		Expect(newlyUpdatedUpgradePathSpecifiers).ShouldNot(ContainElement(upgradePathSpecifier))
	})
})
