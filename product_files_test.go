package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("PivnetClient - product files", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      lager.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = lager.NewLogger("product files test")
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List product files", func() {
		var (
			productSlug string

			response           pivnet.ProductFilesResponse
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"

			response = pivnet.ProductFilesResponse{[]pivnet.ProductFile{
				{
					ID:           1234,
					AWSObjectKey: "something",
				},
				{
					ID:           2345,
					AWSObjectKey: "something-else",
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/product_files",
							apiPrefix,
							productSlug,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			productFiles, err := client.ProductFiles.List(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFiles).To(HaveLen(2))
			Expect(productFiles[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("List product files for release", func() {
		var (
			productSlug string
			releaseID   int

			response           pivnet.ProductFilesResponse
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.ProductFilesResponse{[]pivnet.ProductFile{
				{
					ID:           1234,
					AWSObjectKey: "something",
					Links: &pivnet.Links{Download: map[string]string{
						"href": fmt.Sprintf(
							"/products/%s/releases/%d/product_files/%d/download",
							productSlug,
							releaseID,
							1234,
						)},
					},
				},
				{
					ID:           2345,
					AWSObjectKey: "something-else",
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/product_files",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			productFiles, err := client.ProductFiles.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFiles).To(HaveLen(2))
			Expect(productFiles[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Get Product File", func() {
		var (
			productSlug   string
			releaseID     int
			productFileID int

			response           pivnet.ProductFileResponse
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12
			productFileID = 1234

			response = pivnet.ProductFileResponse{pivnet.ProductFile{
				ID:           productFileID,
				AWSObjectKey: "something",
				Links: &pivnet.Links{Download: map[string]string{
					"href": fmt.Sprintf(
						"/products/%s/releases/%d/product_files/%d/download",
						productSlug,
						releaseID,
						productFileID,
					)},
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/product_files/%d",
							apiPrefix,
							productSlug,
							releaseID,
							productFileID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			productFile, err := client.ProductFiles.Get(
				productSlug,
				releaseID,
				productFileID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(productFileID))
			Expect(productFile.AWSObjectKey).To(Equal("something"))

			Expect(productFile.Links.Download["href"]).
				To(Equal(fmt.Sprintf(
					"/products/%s/releases/%d/product_files/%d/download",
					productSlug,
					releaseID,
					productFileID,
				)))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.Get(
					productSlug,
					releaseID,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Create Product File", func() {
		var (
			createProductFileConfig pivnet.CreateProductFileConfig
		)

		BeforeEach(func() {
			createProductFileConfig = pivnet.CreateProductFileConfig{
				ProductSlug:  productSlug,
				Name:         "some-file-name",
				FileVersion:  "some-file-version",
				AWSObjectKey: "some-aws-object-key",
			}
		})

		Context("when the config is valid", func() {
			type requestBody struct {
				ProductFile pivnet.ProductFile `json:"product_file"`
			}

			const (
				expectedFileType = "Software"
			)

			var (
				expectedRequestBody requestBody

				validResponse = `{"product_file":{"id":1234}}`
			)

			BeforeEach(func() {
				expectedRequestBody = requestBody{
					ProductFile: pivnet.ProductFile{
						FileType:     expectedFileType,
						FileVersion:  createProductFileConfig.FileVersion,
						Name:         createProductFileConfig.Name,
						MD5:          createProductFileConfig.MD5,
						AWSObjectKey: createProductFileConfig.AWSObjectKey,
					},
				}
			})

			It("creates the release with the minimum required fields", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
						ghttp.VerifyJSONRepresenting(&expectedRequestBody),
						ghttp.RespondWith(http.StatusCreated, validResponse),
					),
				)

				productFile, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(productFile.ID).To(Equal(1234))
			})

			Context("when the optional description is present", func() {
				var (
					description string

					productFileResponse pivnet.ProductFileResponse
				)

				BeforeEach(func() {
					description = "some\nmulti-line\ndescription"

					expectedRequestBody.ProductFile.Description = description

					productFileResponse = pivnet.ProductFileResponse{pivnet.ProductFile{
						ID:          1234,
						Description: description,
					}}
				})

				It("creates the product file with the description field", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
							ghttp.VerifyJSONRepresenting(&expectedRequestBody),
							ghttp.RespondWithJSONEncoded(http.StatusCreated, productFileResponse),
						),
					)

					createProductFileConfig.Description = description

					productFile, err := client.ProductFiles.Create(createProductFileConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(productFile.Description).To(Equal(description))
				})
			})
		})

		Context("when the server responds with a non-201 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 201")))
			})
		})

		Context("when the aws object key is empty", func() {
			BeforeEach(func() {
				createProductFileConfig = pivnet.CreateProductFileConfig{
					ProductSlug:  productSlug,
					Name:         "some-file-name",
					FileVersion:  "some-file-version",
					AWSObjectKey: "",
				}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("AWS object key"))
			})
		})
	})

	Describe("Delete Product File", func() {
		var (
			id = 1234
		)

		It("deletes the product file", func() {
			response := []byte(`{"product_file":{"id":1234}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			productFile, err := client.ProductFiles.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(id))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.ProductFiles.Delete(productSlug, id)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Add Product File", func() {
		var (
			productSlug   = "some-product"
			releaseID     = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 204")))
			})
		})
	})

	Describe("Remove Product File", func() {
		var (
			productSlug   = "some-product"
			releaseID     = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.RemoveFromRelease(productSlug, releaseID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.ProductFiles.RemoveFromRelease(productSlug, releaseID, productFileID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 204")))
			})
		})
	})
})
