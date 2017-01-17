package pivnet_test

import (
	"github.com/pivotal-cf/go-pivnet"
	"net/http"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"fmt"
)

var _ = Describe("PivnetClient - ProductFileLinkFetcher", func() {
	var (
		server     *ghttp.Server
		client pivnet.Client
		token string
		pivnetApiAddress string

		newClientConfig pivnet.ClientConfig
		fakeLogger logger.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		pivnetApiAddress = server.URL()
		token = "my-auth-token"

		fakeLogger = &loggerfakes.FakeLogger{}
		newClientConfig = pivnet.ClientConfig{
			Host:      pivnetApiAddress,
			Token:     token,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("NewDownloadLink", func() {
		It("returns a url", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("%s/test-endpoint", apiPrefix)),
					ghttp.RespondWith(http.StatusFound, nil,
						http.Header{
							"Location": []string{"http://example.com"},
						},
					),
				),
			)

			linkFetcher := pivnet.NewProductFileLinkFetcher("/test-endpoint", client)
			link, err := linkFetcher.NewDownloadLink()
			Expect(err).NotTo(HaveOccurred())
			Expect(link).To(Equal("http://example.com"))
		})
	})
})
