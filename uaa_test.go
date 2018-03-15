package pivnet

import (
	"errors"

	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("UAA", func() {
	Describe("TokenFetcher", func() {
		var (
			server       *ghttp.Server
			tokenFetcher *TokenFetcher
		)

		BeforeEach(func() {
			server = ghttp.NewServer()
			tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token")
		})

		AfterEach(func() {
			server.Close()
		})

		It("returns a UAA token without error", func() {
			response := AuthResp{Token: "some-uaa-token"}
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/authentication/access_tokens"),
					ghttp.VerifyBody([]byte(`{"refresh_token":"some-refresh-token"}`)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			token, err := tokenFetcher.GetToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("some-uaa-token"))
		})

		Context("when UAA server responds with a non-200 status code", func() {
			It("returns the error 418", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/authentication/access_tokens"),
						ghttp.VerifyBody([]byte(`{"refresh_token":"some-refresh-token"}`)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, nil),
					),
				)

				_, err := tokenFetcher.GetToken()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(errors.New("failed to fetch API token - received status 418")))
			})

			It("returns an error without endpoint", func() {
				tokenFetcher = NewTokenFetcher("", "some-refresh-token")
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/authentication/access_tokens"),
						ghttp.VerifyBody([]byte(`{"refresh_token":"some-refresh-token"}`)),
						ghttp.RespondWithJSONEncoded(http.StatusTeapot, nil),
					),
				)

				_, err := tokenFetcher.GetToken()
				Expect(err).To(HaveOccurred())
			})
		})

	})
})
