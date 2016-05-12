package commands_test

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
)

var _ = Describe("Pivnet commands", func() {
	var (
		field reflect.StructField
	)

	Describe("redaction", func() {
		var (
			server *ghttp.Server

			outBuffer bytes.Buffer
			client    extension.ExtendedClient
		)

		BeforeEach(func() {
			server = ghttp.NewServer()

			commands.Pivnet.Host = server.URL()
			commands.Pivnet.Verbose = true

			outBuffer = bytes.Buffer{}
			commands.LogWriter = &outBuffer

			client = commands.NewClient()

			products := []pivnet.Product{
				{
					ID:   2345,
					Slug: "another-product-slug",
					Name: "another-product-name",
				},
			}

			productsResponse := pivnet.ProductsResponse{
				Products: products,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, productsResponse),
				),
			)
		})

		It("redacts api token", func() {
			_, err := client.Products.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(outBuffer.String()).Should(ContainSubstring("*** redacted api token ***"))
			Expect(outBuffer.String()).ShouldNot(ContainSubstring(apiToken))
		})

		AfterEach(func() {
			server.Close()

			commands.Pivnet.Verbose = false
		})
	})

	Describe("Version", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "Version")
		})

		It("contains short flag", func() {
			Expect(shortTag(field)).To(Equal("v"))
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("version"))
		})
	})

	Describe("Verbose flag", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "Verbose")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("verbose"))
		})
	})

	Describe("Format flag", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "Format")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("format"))
		})

		It("defaults to table", func() {
			Expect(field.Tag.Get("default")).To(Equal("table"))
		})

		It("contains choice", func() {
			Expect(string(field.Tag)).To(
				MatchRegexp(`choice:"table".*choice:"json".*choice:"yaml"`))
		})

		It("is not required", func() {
			Expect(isRequired(field)).To(BeFalse())
		})
	})

	Describe("APIToken flag", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "APIToken")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("api-token"))
		})

		It("is not required", func() {
			Expect(isRequired(field)).To(BeFalse())
		})
	})

	Describe("Host flag", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "Host")
		})

		It("contains long flag", func() {
			Expect(longTag(field)).To(Equal("host"))
		})

		It("is not required", func() {
			Expect(isRequired(field)).To(BeFalse())
		})
	})

	Describe("ReleaseTypes command", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "ReleaseTypes")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("release-types"))
		})
	})

	Describe("EULAs command", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "EULAs")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("eulas"))
		})
	})

	Describe("EULA command", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "EULA")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("eula"))
		})
	})

	Describe("AcceptEULA command", func() {
		BeforeEach(func() {
			field = fieldFor(commands.Pivnet, "AcceptEULA")
		})

		It("contains command", func() {
			Expect(command(field)).To(Equal("accept-eula"))
		})
	})
})
