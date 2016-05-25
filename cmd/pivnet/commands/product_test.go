package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errors/errorsfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("product commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorsfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		product  pivnet.Product
		products []pivnet.Product

		responseStatusCode int
		response           interface{}
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

		fakeErrorHandler = &errorsfakes.FakeErrorHandler{}
		commands.ErrorHandler = fakeErrorHandler

		product = pivnet.Product{
			ID:   1234,
			Slug: "some-product-slug",
			Name: "some-product-name",
		}

		products = []pivnet.Product{
			product,
			{
				ID:   2345,
				Slug: "another-product-slug",
				Name: "another-product-name",
			},
		}

		responseStatusCode = http.StatusOK
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ProductsCommand", func() {
		var (
			command commands.ProductsCommand
		)

		BeforeEach(func() {
			response = pivnet.ProductsResponse{
				Products: products,
			}

			command = commands.ProductsCommand{}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products", apiPrefix)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("lists all products", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedProducts []pivnet.Product

			err = json.Unmarshal(outBuffer.Bytes(), &returnedProducts)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedProducts).To(Equal(products))
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})
	})

	Describe("ProductCommand", func() {
		var (
			command commands.ProductCommand
		)

		BeforeEach(func() {
			response = product

			command = commands.ProductCommand{
				ProductSlug: product.Slug,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("shows the product", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedProduct pivnet.Product

			err = json.Unmarshal(outBuffer.Bytes(), &returnedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedProduct).To(Equal(product))
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductCommand{}, "ProductSlug")
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
	})
})
