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

var _ = Describe("PivnetClient - helm chart references", func() {
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

	Describe("List helm chart references", func() {
		var (
			productSlug string

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"

			response = pivnet.HelmChartReferencesResponse{[]pivnet.HelmChartReference{
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
							"%s/products/%s/helm_chart_references",
							apiPrefix,
							productSlug,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the helm chart references without error", func() {
			helmChartReferences, err := client.HelmChartReferences.List(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(helmChartReferences).To(HaveLen(2))
			Expect(helmChartReferences[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.HelmChartReferences.List(
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
				_, err := client.HelmChartReferences.List(
					productSlug,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("List helm chart references for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.HelmChartReferencesResponse{[]pivnet.HelmChartReference{
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
							"%s/products/%s/releases/%d/helm_chart_references",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the helm chart references without error", func() {
			helmChartReferences, err := client.HelmChartReferences.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(helmChartReferences).To(HaveLen(2))
			Expect(helmChartReferences[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.HelmChartReferences.ListForRelease(
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
				_, err := client.HelmChartReferences.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get Helm Chart Reference", func() {
		var (
			productSlug      string
			helmChartReferenceID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			helmChartReferenceID = 1234

			response = pivnet.HelmChartReferenceResponse{
				HelmChartReference: pivnet.HelmChartReference{
					ID:   helmChartReferenceID,
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
							"%s/products/%s/helm_chart_references/%d",
							apiPrefix,
							productSlug,
							helmChartReferenceID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the helm chart reference without error", func() {
			helmChartReference, err := client.HelmChartReferences.Get(
				productSlug,
				helmChartReferenceID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(helmChartReference.ID).To(Equal(helmChartReferenceID))
			Expect(helmChartReference.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.HelmChartReferences.Get(
					productSlug,
					helmChartReferenceID,
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
				_, err := client.HelmChartReferences.Get(
					productSlug,
					helmChartReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Get helm chart reference for release", func() {
		var (
			productSlug      string
			releaseID        int
			helmChartReferenceID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12
			helmChartReferenceID = 1234

			response = pivnet.HelmChartReferenceResponse{
				HelmChartReference: pivnet.HelmChartReference{
					ID:   helmChartReferenceID,
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
							"%s/products/%s/releases/%d/helm_chart_references/%d",
							apiPrefix,
							productSlug,
							releaseID,
							helmChartReferenceID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the helm chart reference without error", func() {
			helmChartReference, err := client.HelmChartReferences.GetForRelease(
				productSlug,
				releaseID,
				helmChartReferenceID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(helmChartReference.ID).To(Equal(helmChartReferenceID))
			Expect(helmChartReference.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.HelmChartReferences.GetForRelease(
					productSlug,
					releaseID,
					helmChartReferenceID,
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
				_, err := client.HelmChartReferences.GetForRelease(
					productSlug,
					releaseID,
					helmChartReferenceID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("json"))
			})
		})
	})

	Describe("Create Helm Chart Reference", func() {
		type requestBody struct {
			HelmChartReference pivnet.HelmChartReference `json:"helm_chart_reference"`
		}

		var (
			createHelmChartReferenceConfig pivnet.CreateHelmChartReferenceConfig

			expectedRequestBody requestBody

			helmChartReferenceResponse pivnet.HelmChartReferenceResponse
		)

		BeforeEach(func() {
			createHelmChartReferenceConfig = pivnet.CreateHelmChartReferenceConfig{
				ProductSlug:        productSlug,
				Description:        "some\nmulti-line\ndescription",
				DocsURL:            "some-docs-url",
				Name:               "some-helm-chart-name",
				Version:            "1.2.3",
				SystemRequirements: []string{"system-1", "system-2"},
			}

			expectedRequestBody = requestBody{
				HelmChartReference: pivnet.HelmChartReference{
					Description:        createHelmChartReferenceConfig.Description,
					DocsURL:            createHelmChartReferenceConfig.DocsURL,
					Name:               createHelmChartReferenceConfig.Name,
					Version:            createHelmChartReferenceConfig.Version,
					SystemRequirements: createHelmChartReferenceConfig.SystemRequirements,
				},
			}

			helmChartReferenceResponse = pivnet.HelmChartReferenceResponse{
				HelmChartReference: pivnet.HelmChartReference{
					ID:                 1234,
					Description:        createHelmChartReferenceConfig.Description,
					DocsURL:            createHelmChartReferenceConfig.DocsURL,
					Name:               createHelmChartReferenceConfig.Name,
					Version:            createHelmChartReferenceConfig.Version,
					SystemRequirements: createHelmChartReferenceConfig.SystemRequirements,
				}}
		})

		It("creates the helm chart reference", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/helm_chart_references",
						apiPrefix,
						productSlug,
					)),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, helmChartReferenceResponse),
				),
			)

			helmChartReference, err := client.HelmChartReferences.Create(createHelmChartReferenceConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(helmChartReference.ID).To(Equal(1234))
			Expect(helmChartReference).To(Equal(helmChartReferenceResponse.HelmChartReference))
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
							"%s/products/%s/helm_chart_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.HelmChartReferences.Create(createHelmChartReferenceConfig)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the server responds with a 429 status code", func() {
			It("returns an error indicating the limit was hit", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/helm_chart_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTooManyRequests, "Retry later"),
					),
				)

				_, err := client.HelmChartReferences.Create(createHelmChartReferenceConfig)
				Expect(err.Error()).To(ContainSubstring("You have hit the helm chart reference creation limit. Please wait before creating more helm chart references. Contact pivnet-eng@pivotal.io with additional questions."))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/helm_chart_references",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.HelmChartReferences.Create(createHelmChartReferenceConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Update helm chart reference", func() {
		type requestBody struct {
			HelmChartReference pivnet.HelmChartReference `json:"helm_chart_reference"`
		}

		var (
			expectedRequestBody     requestBody
			helmChartReference          pivnet.HelmChartReference
			updateHelmChartReferenceUrl string
			validResponse           = `{"helm_chart_reference":{"id":1234, "docs_url":"example.io", "system_requirements": ["1", "2"]}}`
		)

		BeforeEach(func() {
			helmChartReference = pivnet.HelmChartReference{
				ID:                 1234,
				Description:        "Avast! Pieces o' passion are forever fine.",
				DocsURL:            "example.io",
				Name:               "turpis-hercle",
				Version:            "1.2.3",
				SystemRequirements: []string{"1", "2"},
			}

			expectedRequestBody = requestBody{
				HelmChartReference: pivnet.HelmChartReference{
					Description:        helmChartReference.Description,
					Name:               helmChartReference.Name,
					DocsURL:            helmChartReference.DocsURL,
					SystemRequirements: helmChartReference.SystemRequirements,
				},
			}

			updateHelmChartReferenceUrl = fmt.Sprintf(
				"%s/products/%s/helm_chart_references/%d",
				apiPrefix,
				productSlug,
				helmChartReference.ID,
			)

		})

		It("updates the helm chart reference with the provided fields", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateHelmChartReferenceUrl),
					ghttp.VerifyJSONRepresenting(&expectedRequestBody),
					ghttp.RespondWith(http.StatusOK, validResponse),
				),
			)

			updatedHelmChartReference, err := client.HelmChartReferences.Update(productSlug, helmChartReference)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedHelmChartReference.ID).To(Equal(helmChartReference.ID))
			Expect(updatedHelmChartReference.DocsURL).To(Equal(helmChartReference.DocsURL))
			Expect(updatedHelmChartReference.SystemRequirements).To(ConsistOf("2", "1"))
		})

		It("forwards the server-side error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateHelmChartReferenceUrl),
					ghttp.RespondWithJSONEncoded(http.StatusTeapot,
						pivnetErr{Message: "Meet, scotty, powerdrain!"}),
				),
			)

			_, err := client.HelmChartReferences.Update(productSlug, helmChartReference)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("scotty"))
		})

		It("forwards the unmarshalling error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", updateHelmChartReferenceUrl),
					ghttp.RespondWith(http.StatusTeapot, "<NOT></JSON>"),
				),
			)
			_, err := client.HelmChartReferences.Update(productSlug, helmChartReference)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})

	Describe("Delete Helm Chart Reference", func() {
		var (
			id = 1234
		)

		It("deletes the helm chart reference", func() {
			response := []byte(`{"helm_chart_reference":{"id":1234}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/helm_chart_references/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			helmChartReference, err := client.HelmChartReferences.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(helmChartReference.ID).To(Equal(id))
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
							fmt.Sprintf("%s/products/%s/helm_chart_references/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				_, err := client.HelmChartReferences.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/helm_chart_references/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.HelmChartReferences.Delete(productSlug, id)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Add Helm Chart Reference to release", func() {
		var (
			productSlug      = "some-product"
			releaseID        = 2345
			helmChartReferenceID = 3456

			expectedRequestBody = `{"helm_chart_reference":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_helm_chart_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.HelmChartReferences.AddToRelease(productSlug, releaseID, helmChartReferenceID)
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
							"%s/products/%s/releases/%d/add_helm_chart_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.HelmChartReferences.AddToRelease(productSlug, releaseID, helmChartReferenceID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_helm_chart_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.HelmChartReferences.AddToRelease(productSlug, releaseID, helmChartReferenceID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})

	Describe("Remove Helm Chart Reference from release", func() {
		var (
			productSlug      = "some-product"
			releaseID        = 2345
			helmChartReferenceID = 3456

			expectedRequestBody = `{"helm_chart_reference":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_helm_chart_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.HelmChartReferences.RemoveFromRelease(productSlug, releaseID, helmChartReferenceID)
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
							"%s/products/%s/releases/%d/remove_helm_chart_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, response),
					),
				)

				err := client.HelmChartReferences.RemoveFromRelease(productSlug, releaseID, helmChartReferenceID)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_helm_chart_reference",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				err := client.HelmChartReferences.RemoveFromRelease(productSlug, releaseID, helmChartReferenceID)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
