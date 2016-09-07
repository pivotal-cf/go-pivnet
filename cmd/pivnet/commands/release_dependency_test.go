package commands_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/commandsfakes"
)

var _ = Describe("release dependency commands", func() {
	var (
		field reflect.StructField

		fakeReleaseDependencyClient *commandsfakes.FakeReleaseDependencyClient
	)

	BeforeEach(func() {
		fakeReleaseDependencyClient = &commandsfakes.FakeReleaseDependencyClient{}

		commands.NewReleaseDependencyClient = func() commands.ReleaseDependencyClient {
			return fakeReleaseDependencyClient
		}
	})

	Describe("ReleasesDependenciesCommand", func() {
		var (
			cmd *commands.ReleaseDependenciesCommand
		)

		BeforeEach(func() {
			cmd = &commands.ReleaseDependenciesCommand{}
		})

		It("invokes the ReleaseDependency client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseDependencyClient.ListCallCount()).To(Equal(1))
		})

		Context("when the ReleaseDependency client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeReleaseDependencyClient.ListReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseDependenciesCommand{}, "ProductSlug")
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
				field = fieldFor(commands.ReleaseDependenciesCommand{}, "ReleaseVersion")
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
