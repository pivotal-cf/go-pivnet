package pivnet_test

import (
	"fmt"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - Auth", func() {
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

	Describe("Check", func() {
		It("returns true,nil", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/authentication", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)

			ok, err := client.Auth.Check()
			Expect(err).NotTo(HaveOccurred())

			Expect(ok).To(BeTrue())
		})

		Context("when the server responds with a 401 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns false,nil", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/authentication", apiPrefix)),
						ghttp.RespondWith(http.StatusUnauthorized, body),
					),
				)

				ok, err := client.Auth.Check()
				Expect(err).NotTo(HaveOccurred())

				Expect(ok).To(BeFalse())
			})
		})

		Context("when the server responds with a 403 status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns false,nil", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/authentication", apiPrefix)),
						ghttp.RespondWith(http.StatusForbidden, body),
					),
				)

				ok, err := client.Auth.Check()
				Expect(err).NotTo(HaveOccurred())

				Expect(ok).To(BeFalse())
			})
		})

		Context("when the server responds with any other status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns false,err", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/authentication", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				ok, err := client.Auth.Check()
				Expect(err.Error()).To(ContainSubstring("foo message"))

				Expect(ok).To(BeFalse())
			})
		})
	})
})
