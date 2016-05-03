package commands_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
)

var _ = Describe("Pivnet commands", func() {
	var (
		field reflect.StructField
	)

	var fieldFor = func(name string) reflect.StructField {
		field, success := reflect.TypeOf(commands.Pivnet).FieldByName(name)
		Expect(success).To(BeTrue())
		return field
	}

	var longTag = func(f reflect.StructField) string {
		return f.Tag.Get("long")
	}

	var shortTag = func(f reflect.StructField) string {
		return f.Tag.Get("short")
	}

	var command = func(f reflect.StructField) string {
		return f.Tag.Get("command")
	}

	Describe("Version", func() {
		BeforeEach(func() {
			field = fieldFor("Version")
		})

		It("contains short flag", func() {
			Expect(shortTag(field)).To(Equal("v"))
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("version"))
		})
	})

	Describe("Format", func() {
		BeforeEach(func() {
			field = fieldFor("Format")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("format"))
		})

		It("defaults to table", func() {
			Expect(field.Tag.Get("default")).To(Equal("table"))
		})

		It("contains choice", func() {
			Expect(string(field.Tag)).To(MatchRegexp(`choice:"table".*choice:"json".*choice:"yaml"`))
		})
	})

	Describe("APIToken", func() {
		BeforeEach(func() {
			field = fieldFor("APIToken")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("api-token"))
		})
	})

	Describe("Host", func() {
		BeforeEach(func() {
			field = fieldFor("Host")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("host"))
		})
	})

	Describe("ReleaseTypes", func() {
		BeforeEach(func() {
			field = fieldFor("ReleaseTypes")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("release-types"))
		})
	})

	Describe("EULAs", func() {
		BeforeEach(func() {
			field = fieldFor("EULAs")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("eulas"))
		})
	})

	Describe("EULA", func() {
		BeforeEach(func() {
			field = fieldFor("EULA")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("eula"))
		})
	})

	Describe("AcceptEULA", func() {
		BeforeEach(func() {
			field = fieldFor("AcceptEULA")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("accept-eula"))
		})
	})
})
