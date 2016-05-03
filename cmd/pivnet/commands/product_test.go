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

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("product commands", func() {
	var (
		server *ghttp.Server

		field     reflect.StructField
		outBuffer bytes.Buffer

		product  pivnet.Product
		products []pivnet.Product
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutWriter = &outBuffer

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
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ProductsCommand", func() {
		It("lists all products", func() {
			productsResponse := pivnet.ProductsResponse{
				Products: products,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, productsResponse),
				),
			)

			productsCommand := commands.ProductsCommand{}

			err := productsCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedProducts []pivnet.Product

			err = json.Unmarshal(outBuffer.Bytes(), &returnedProducts)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedProducts).To(Equal(products))
		})
	})

	Describe("ProductCommand", func() {
		It("shows product", func() {
			productsResponse := product

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, productsResponse),
				),
			)

			productCommand := commands.ProductCommand{}
			productCommand.ProductSlug = product.Slug

			err := productCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedProduct pivnet.Product

			err = json.Unmarshal(outBuffer.Bytes(), &returnedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedProduct).To(Equal(product))
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})
	})
})
