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

var _ = Describe("PivnetClient - release upgrade paths", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      lager.Logger

		productID int
		releaseID int
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		productID = 1234
		releaseID = 2345

		fakeLogger = lager.NewLogger("release upgrade paths")
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
		It("returns the release upgrade paths", func() {

			response := pivnet.ReleaseUpgradePathsResponse{
				ReleaseUpgradePaths: []pivnet.ReleaseUpgradePath{
					{
						Release: pivnet.UpgradePathRelease{
							ID:      9876,
							Version: "release 9876",
						},
					},
					{
						Release: pivnet.UpgradePathRelease{
							ID:      8765,
							Version: "release 8765",
						},
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/upgrade_paths",
						apiPrefix,
						productSlug,
						releaseID,
					)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			releaseUpgradePaths, err := client.ReleaseUpgradePaths.Get(productSlug, releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(releaseUpgradePaths).To(HaveLen(2))
			Expect(releaseUpgradePaths[0].Release.ID).To(Equal(9876))
			Expect(releaseUpgradePaths[1].Release.ID).To(Equal(8765))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_paths",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)
			})

			It("returns an error", func() {
				_, err := client.ReleaseUpgradePaths.Get(productSlug, releaseID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})
})
