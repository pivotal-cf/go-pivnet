package commands_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/commandsfakes"
)

var _ = Describe("release upgrade path commands", func() {
	var (
		field reflect.StructField

		fakeReleaseUpgradePathClient *commandsfakes.FakeReleaseUpgradePathClient
	)

	BeforeEach(func() {
		fakeReleaseUpgradePathClient = &commandsfakes.FakeReleaseUpgradePathClient{}

		commands.NewReleaseUpgradePathClient = func() commands.ReleaseUpgradePathClient {
			return fakeReleaseUpgradePathClient
		}
	})

	Describe("ReleasesUpgradePathsCommand", func() {
		var (
			cmd commands.ReleaseUpgradePathsCommand
		)

		BeforeEach(func() {
			cmd = commands.ReleaseUpgradePathsCommand{}
		})

		It("invokes the ReleaseUpgradePath client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseUpgradePathClient.ListCallCount()).To(Equal(1))
		})

		Context("when the ReleaseUpgradePath client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeReleaseUpgradePathClient.ListReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseUpgradePathsCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseUpgradePathsCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("r"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})
})
