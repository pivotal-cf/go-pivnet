package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-cf-experimental/go-pivnet/logger/loggerfakes"
)

var _ = Describe("PivnetClient - release dependencies", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger

		productSlug string
		releaseID   int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		productSlug = "some-product"
		releaseID = 2345

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

	Describe("Get", func() {
		It("returns the release dependencies", func() {

			response := pivnet.ReleaseDependenciesResponse{
				ReleaseDependencies: []pivnet.ReleaseDependency{
					{
						Release: pivnet.DependentRelease{
							ID:      9876,
							Version: "release 9876",
							Product: pivnet.Product{
								ID:   23,
								Name: "Product 23",
							},
						},
					},
					{
						Release: pivnet.DependentRelease{
							ID:      8765,
							Version: "release 8765",
							Product: pivnet.Product{
								ID:   23,
								Name: "Product 23",
							},
						},
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/dependencies",
						apiPrefix,
						productSlug,
						releaseID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			releaseDependencies, err := client.ReleaseDependencies.List(productSlug, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(releaseDependencies).To(HaveLen(2))
			Expect(releaseDependencies[0].Release.ID).To(Equal(9876))
			Expect(releaseDependencies[1].Release.ID).To(Equal(8765))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/dependencies",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.ReleaseDependencies.List(productSlug, releaseID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})
})
