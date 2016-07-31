package commands_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
)

var _ = Describe("eula commands", func() {
	var (
		field reflect.StructField
	)

	Describe("EULACommand", func() {
		Describe("EULASlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.EULACommand{}, "EULASlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("eula-slug"))
			})
		})
	})

	Describe("AcceptEULACommand", func() {
		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AcceptEULACommand{}, "ProductSlug")
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
				field = fieldFor(commands.AcceptEULACommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})
})
