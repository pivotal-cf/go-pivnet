package pivnet_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-cf-experimental/go-pivnet/logger/loggerfakes"
)

var _ = Describe("PivnetClient - release types", func() {
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

	Describe("Get", func() {
		It("returns the release types", func() {
			response := `{"release_types": ["foo","bar"]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			releaseTypes, err := client.ReleaseTypes.Get()
			Expect(err).NotTo(HaveOccurred())

			Expect(releaseTypes).To(HaveLen(2))
			Expect(releaseTypes[0]).To(Equal(pivnet.ReleaseType("foo")))
			Expect(releaseTypes[1]).To(Equal(pivnet.ReleaseType("bar")))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.ReleaseTypes.Get()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})
})
