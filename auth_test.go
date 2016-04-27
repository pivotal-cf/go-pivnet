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

var _ = Describe("PivnetClient - Auth", func() {
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

		fakeLogger = lager.NewLogger("eula test")
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

	Describe("Check", func() {
		It("returns successfully", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/authentication", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)

			err := client.Auth.Check()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/authentication", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.Auth.Check()
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Accept", func() {
		var (
			releaseID         int
			productSlug       string
			EULAAcceptanceURL string
		)

		BeforeEach(func() {
			productSlug = "banana-slug"
			releaseID = 42
			EULAAcceptanceURL = fmt.Sprintf(apiPrefix+"/products/%s/releases/%d/eula_acceptance", productSlug, releaseID)
		})

		It("accepts the EULA for a given release and product ID", func() {
			response := fmt.Sprintf(`{"accepted_at": "2016-01-11","_links":{}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", EULAAcceptanceURL),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.VerifyJSON(`{}`),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			Expect(client.EULA.Accept(productSlug, releaseID)).To(Succeed())
		})

		Context("when any other non-200 status code comes back", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", EULAAcceptanceURL),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
						ghttp.VerifyJSON(`{}`),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				Expect(client.EULA.Accept(productSlug, releaseID)).To(MatchError("Pivnet returned status code: 418 for the request - expected 200"))
			})
		})
	})
})
