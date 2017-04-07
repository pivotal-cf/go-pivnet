package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf/go-pivnet"
)

const (
	dependentProductSlug = "stemcells"
	dependencySpecifier  = "3312.*"
)

var _ = Describe("Dependency Specifier Integration", func() {
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

	It("creates, lists, and deletes dependency specifiers", func() {
		By("Listing dependency specifiers")
		dependencySpecifiers, err := client.DependencySpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

		By("Creating new dependency specifier")
		dependencySpecifier, err := client.DependencySpecifiers.Create(
			testProductSlug,
			newRelease.ID,
			dependentProductSlug,
			dependencySpecifier,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(dependencySpecifiers).ShouldNot(ContainElement(dependencySpecifier))

		By("Re-listing dependency specifiers")
		updatedDependencySpecifiers, err := client.DependencySpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

		Expect(updatedDependencySpecifiers).Should(ContainElement(dependencySpecifier))

		By("Getting individual dependency specifier")
		individualDependencySpecifier, err := client.DependencySpecifiers.Get(
			testProductSlug,
			newRelease.ID,
			dependencySpecifier.ID,
		)
		Expect(err).NotTo(HaveOccurred())

		Expect(individualDependencySpecifier).To(Equal(dependencySpecifier))

		By("Deleting dependency specifier")
		err = client.DependencySpecifiers.Delete(
			testProductSlug,
			newRelease.ID,
			dependencySpecifier.ID,
		)
		Expect(err).NotTo(HaveOccurred())

		newlyUpdatedDependencySpecifiers, err := client.DependencySpecifiers.List(testProductSlug, newRelease.ID)
		Expect(err).NotTo(HaveOccurred())

		Expect(newlyUpdatedDependencySpecifiers).ShouldNot(ContainElement(dependencySpecifier))
	})
})
