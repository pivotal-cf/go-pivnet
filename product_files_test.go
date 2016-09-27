package pivnet_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
)

var _ = Describe("PivnetClient - product files", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
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

			response           interface{}
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

		It("returns the product files without error", func() {
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
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("List product files for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
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

		It("returns the product files without error", func() {
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
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Get Product File", func() {
		var (
			productSlug   string
			productFileID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			productFileID = 1234

			response = pivnet.ProductFileResponse{
				ProductFile: pivnet.ProductFile{
					ID:           productFileID,
					AWSObjectKey: "something",
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							productSlug,
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
				productFileID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(productFileID))
			Expect(productFile.AWSObjectKey).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.Get(
					productSlug,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Get product file for release", func() {
		var (
			productSlug   string
			releaseID     int
			productFileID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12
			productFileID = 1234

			response = pivnet.ProductFileResponse{
				ProductFile: pivnet.ProductFile{
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
			productFile, err := client.ProductFiles.GetForRelease(
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
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ProductFiles.GetForRelease(
					productSlug,
					releaseID,
					productFileID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
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
				FileType:     "some-file-type",
			}
		})

		Context("when the config is valid", func() {
			type requestBody struct {
				ProductFile pivnet.ProductFile `json:"product_file"`
			}

			var (
				expectedRequestBody requestBody

				validResponse = `{"product_file":{"id":1234}}`
			)

			BeforeEach(func() {
				expectedRequestBody = requestBody{
					ProductFile: pivnet.ProductFile{
						FileType:     "some-file-type",
						FileVersion:  createProductFileConfig.FileVersion,
						Name:         createProductFileConfig.Name,
						MD5:          createProductFileConfig.MD5,
						AWSObjectKey: createProductFileConfig.AWSObjectKey,
					},
				}
			})

			It("creates the product file with the minimum required fields", func() {
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

					productFileResponse = pivnet.ProductFileResponse{
						ProductFile: pivnet.ProductFile{
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
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ProductFiles.Create(createProductFileConfig)
				Expect(err.Error()).To(ContainSubstring("foo message"))
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

	Describe("Update Product File", func() {
		type requestBody struct {
			ProductFile pivnet.ProductFile `json:"product_file"`
		}

		var (
			expectedRequestBody requestBody

			productFile pivnet.ProductFile

			validResponse = `{"product_file":{"id":1234}}`
		)

		BeforeEach(func() {
			productFile = pivnet.ProductFile{
				ID:          1234,
				Description: "some-description",
				FileVersion: "some-file-version",
				FileType:    "some-file-type",
				MD5:         "some-md5",
				Name:        "some-file-name",
			}

			expectedRequestBody = requestBody{
				ProductFile: pivnet.ProductFile{
					Description: productFile.Description,
					FileType:    productFile.FileType,
					FileVersion: productFile.FileVersion,
					MD5:         productFile.MD5,
					Name:        productFile.Name,
				},
			}
		})

		It("updates the product file with the provided fields", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf(
						"%s/products/%s/product_files/%d",
						apiPrefix,
						productSlug,
						productFile.ID,
					)),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWith(http.StatusOK, validResponse),
				),
			)

			updatedProductFile, err := client.ProductFiles.Update(productSlug, productFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedProductFile.ID).To(Equal(productFile.ID))
		})

		Context("when the server responds with a non-200 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							productSlug,
							productFile.ID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ProductFiles.Update(productSlug, productFile)
				Expect(err.Error()).To(ContainSubstring("foo message"))
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
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ProductFiles.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Add Product File to release", func() {
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
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.AddToRelease(productSlug, releaseID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Remove Product File from release", func() {
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
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.RemoveFromRelease(productSlug, releaseID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Add Product File to file group", func() {
		var (
			productSlug   = "some-product"
			fileGroupID   = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/add_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.AddToFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/add_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.AddToFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Remove Product File from file group", func() {
		var (
			productSlug   = "some-product"
			fileGroupID   = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/remove_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ProductFiles.RemoveFromFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			var (
				response interface{}
			)

			BeforeEach(func() {
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/file_groups/%d/remove_product_file",
							apiPrefix,
							productSlug,
							fileGroupID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ProductFiles.RemoveFromFileGroup(productSlug, fileGroupID, productFileID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})
})
