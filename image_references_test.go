package pivnet_test

import (
	"fmt"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v3"
	"github.com/pivotal-cf/go-pivnet/v3/go-pivnetfakes"
	"github.com/pivotal-cf/go-pivnet/v3/logger"
	"github.com/pivotal-cf/go-pivnet/v3/logger/loggerfakes"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - image references", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List image references", func() {
		var (
			productSlug string

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"

			response = pivnet.ImageReferencesResponse{[]pivnet.ImageReference{
				{
					ID:   1234,
					Name: "something",
				},
				{
					ID:   2345,
					Name: "something-else",
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
							"%s/products/%s/image_references",
							apiPrefix,
							productSlug,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the image references without error", func() {
			imageReferences, err := client.ImageReferences.List(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(imageReferences).To(HaveLen(2))
			Expect(imageReferences[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ImageReferences.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ImageReferences.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("List image references for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.ImageReferencesResponse{[]pivnet.ImageReference{
				{
					ID:   1234,
					Name: "something",
				},
				{
					ID:   2345,
					Name: "something-else",
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
							"%s/products/%s/releases/%d/image_references",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the image references without error", func() {
			imageReferences, err := client.ImageReferences.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(imageReferences).To(HaveLen(2))
			Expect(imageReferences[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ImageReferences.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ImageReferences.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get Image Reference", func() {
		var (
			productSlug      string
			imageReferenceID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			imageReferenceID = 1234

			response = pivnet.ImageReferenceResponse{
				ImageReference: pivnet.ImageReference{
					ID:   imageReferenceID,
					Name: "something",
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/image_references/%d",
							apiPrefix,
							productSlug,
							imageReferenceID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the image reference without error", func() {
			imageReference, err := client.ImageReferences.Get(
				productSlug,
				imageReferenceID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(imageReference.ID).To(Equal(imageReferenceID))
			Expect(imageReference.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ImageReferences.Get(
					productSlug,
					imageReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ImageReferences.Get(
					productSlug,
					imageReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get image reference for release", func() {
		var (
			productSlug      string
			releaseID        int
			imageReferenceID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12
			imageReferenceID = 1234

			response = pivnet.ImageReferenceResponse{
				ImageReference: pivnet.ImageReference{
					ID:   imageReferenceID,
					Name: "something",
				}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/image_references/%d",
							apiPrefix,
							productSlug,
							releaseID,
							imageReferenceID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the image reference without error", func() {
			imageReference, err := client.ImageReferences.GetForRelease(
				productSlug,
				releaseID,
				imageReferenceID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(imageReference.ID).To(Equal(imageReferenceID))
			Expect(imageReference.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.ImageReferences.GetForRelease(
					productSlug,
					releaseID,
					imageReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			BeforeEach(func() {
				response = "%%%"
			})

			It("forwards the error", func() {
				_, err := client.ImageReferences.GetForRelease(
					productSlug,
					releaseID,
					imageReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Create Image Reference", func() {
		type requestBody struct {
			ImageReference pivnet.ImageReference `json:"image_reference"`
		}

		var (
			createImageReferenceConfig pivnet.CreateImageReferenceConfig

			expectedRequestBody requestBody

			imageReferenceResponse pivnet.ImageReferenceResponse
		)

		BeforeEach(func() {
			createImageReferenceConfig = pivnet.CreateImageReferenceConfig{
				ProductSlug:        productSlug,
				Description:        "some\nmulti-line\ndescription",
				Digest:             "sha256:mydigest",
				DocsURL:            "some-docs-url",
				ImagePath:          "my/path:123",
				Name:               "some-image-name",
				SystemRequirements: []string{"system-1", "system-2"},
			}

			expectedRequestBody = requestBody{
				ImageReference: pivnet.ImageReference{
					Description:        createImageReferenceConfig.Description,
					Digest:             createImageReferenceConfig.Digest,
					DocsURL:            createImageReferenceConfig.DocsURL,
					ImagePath:          createImageReferenceConfig.ImagePath,
					Name:               createImageReferenceConfig.Name,
					SystemRequirements: createImageReferenceConfig.SystemRequirements,
				},
			}

			imageReferenceResponse = pivnet.ImageReferenceResponse{
				ImageReference: pivnet.ImageReference{
					ID:                 1234,
					Description:        createImageReferenceConfig.Description,
					Digest:             createImageReferenceConfig.Digest,
					DocsURL:            createImageReferenceConfig.DocsURL,
					ImagePath:          createImageReferenceConfig.ImagePath,
					Name:               createImageReferenceConfig.Name,
					SystemRequirements: createImageReferenceConfig.SystemRequirements,
				}}
		})

		It("creates the image reference", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/image_references",
						apiPrefix,
						productSlug,
					)),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, imageReferenceResponse),
				),
			)

			imageReference, err := client.ImageReferences.Create(createImageReferenceConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageReference.ID).To(Equal(1234))
			Expect(imageReference).To(Equal(imageReferenceResponse.ImageReference))
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
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/image_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ImageReferences.Create(createImageReferenceConfig)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the server responds with a 429 status code", func() {
			It("returns an error indicating the limit was hit", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/image_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTooManyRequests, "Retry later"),
					),
				)

				_, err := client.ImageReferences.Create(createImageReferenceConfig)
				Expect(err.Error()).To(ContainSubstring("You have hit the image reference creation limit. Please wait before creating more image references. Contact pivnet-eng@pivotal.io with additional questions."))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/image_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ImageReferences.Create(createImageReferenceConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Update image reference", func() {
		type requestBody struct {
			ImageReference pivnet.ImageReference `json:"image_reference"`
		}

		var (
			expectedRequestBody     requestBody
			imageReference          pivnet.ImageReference
			updateImageReferenceUrl string
			validResponse           = `{"image_reference":{"id":1234, "docs_url":"example.io", "system_requirements": ["1", "2"]}}`
		)

		BeforeEach(func() {
			imageReference = pivnet.ImageReference{
				ID:                 1234,
				ImagePath:          "some/path",
				Description:        "Avast! Pieces o' passion are forever fine.",
				Digest:             "some-sha265",
				DocsURL:            "example.io",
				Name:               "turpis-hercle",
				SystemRequirements: []string{"1", "2"},
			}

			expectedRequestBody = requestBody{
				ImageReference: pivnet.ImageReference{
					Description:        imageReference.Description,
					Name:               imageReference.Name,
					DocsURL:            imageReference.DocsURL,
					SystemRequirements: imageReference.SystemRequirements,
				},
			}

			updateImageReferenceUrl = fmt.Sprintf(
				"%s/products/%s/image_references/%d",
				apiPrefix,
				productSlug,
				imageReference.ID,
			)

		})

		It("updates the image reference with the provided fields", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateImageReferenceUrl),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWith(http.StatusOK, validResponse),
				),
			)

			updatedImageReference, err := client.ImageReferences.Update(productSlug, imageReference)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedImageReference.ID).To(Equal(imageReference.ID))
			Expect(updatedImageReference.DocsURL).To(Equal(imageReference.DocsURL))
			Expect(updatedImageReference.SystemRequirements).To(ConsistOf("2", "1"))
		})

		It("forwards the server-side error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateImageReferenceUrl),
					ghttp.RespondWithJSONEncoded(http.StatusTeapot,
						pivnetErr{Message: "Meet, scotty, powerdrain!"}),
				),
			)

			_, err := client.ImageReferences.Update(productSlug, imageReference)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("scotty"))
		})

		It("forwards the unmarshalling error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateImageReferenceUrl),
					ghttp.RespondWith(http.StatusTeapot, "<NOT></JSON>"),
				),
			)
			_, err := client.ImageReferences.Update(productSlug, imageReference)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})

	Describe("Delete Image Reference", func() {
		var (
			id = 1234
		)

		It("deletes the image reference", func() {
			response := []byte(`{"image_reference":{"id":1234}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/image_references/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			imageReference, err := client.ImageReferences.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(imageReference.ID).To(Equal(id))
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
							fmt.Sprintf("%s/products/%s/image_references/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.ImageReferences.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/image_references/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.ImageReferences.Delete(productSlug, id)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add Image Reference to release", func() {
		var (
			productSlug      = "some-product"
			releaseID        = 2345
			imageReferenceID = 3456

			expectedRequestBody = `{"image_reference":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_image_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ImageReferences.AddToRelease(productSlug, releaseID, imageReferenceID)
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
							"%s/products/%s/releases/%d/add_image_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ImageReferences.AddToRelease(productSlug, releaseID, imageReferenceID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_image_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ImageReferences.AddToRelease(productSlug, releaseID, imageReferenceID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove Image Reference from release", func() {
		var (
			productSlug      = "some-product"
			releaseID        = 2345
			imageReferenceID = 3456

			expectedRequestBody = `{"image_reference":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_image_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.ImageReferences.RemoveFromRelease(productSlug, releaseID, imageReferenceID)
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
							"%s/products/%s/releases/%d/remove_image_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.ImageReferences.RemoveFromRelease(productSlug, releaseID, imageReferenceID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_image_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.ImageReferences.RemoveFromRelease(productSlug, releaseID, imageReferenceID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
