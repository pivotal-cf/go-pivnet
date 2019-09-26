package pivnet_test

import (
	"fmt"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v2"
	"github.com/pivotal-cf/go-pivnet/v2/go-pivnetfakes"
	"github.com/pivotal-cf/go-pivnet/v2/logger"
	"github.com/pivotal-cf/go-pivnet/v2/logger/loggerfakes"
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
					ID:           1234,
					ImagePath: "something",
				},
				{
					ID:           2345,
					ImagePath: "something-else",
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

		It("returns the product files without error", func() {
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
					ID:           1234,
					ImagePath: "something",
				},
				{
					ID:           2345,
					ImagePath: "something-else",
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

})
