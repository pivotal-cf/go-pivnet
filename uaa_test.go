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
			tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", ProxyAuthConfig{})
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

		It("passes on the user agent in the request header", func() {
			userAgent := "my_user_agent"
			tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, userAgent, ProxyAuthConfig{})
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("User-Agent", userAgent),
				),
			)

			tokenFetcher.GetToken()
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
				tokenFetcher = NewTokenFetcher("", "some-refresh-token", false, "", ProxyAuthConfig{})
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

		Context("when proxy authentication is configured", func() {
			Context("with Basic authentication", func() {
				It("returns an error when proxy URL is empty but auth type is set", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "user",
						Password: "pass",
						ProxyURL: "",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					_, err := tokenFetcher.GetToken()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("proxy URL is required"))
				})

				It("returns an error when proxy URL is invalid", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "user",
						Password: "pass",
						ProxyURL: "://invalid-url",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					_, err := tokenFetcher.GetToken()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to parse proxy URL"))
				})

				It("accepts valid proxy auth config without error", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "proxyuser",
						Password: "proxypass",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					// TokenFetcher should be created successfully with proxy config
					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.ProxyAuthConfig.AuthType).To(Equal(ProxyAuthTypeBasic))
					Expect(tokenFetcher.ProxyAuthConfig.Username).To(Equal("proxyuser"))
				})

				It("accepts empty username and password for Basic auth", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "",
						Password: "",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.ProxyAuthConfig.AuthType).To(Equal(ProxyAuthTypeBasic))
				})

				It("handles special characters in username and password", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "user@domain.com",
						Password: "p@$$w0rd!#%",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.ProxyAuthConfig.Username).To(Equal("user@domain.com"))
					Expect(tokenFetcher.ProxyAuthConfig.Password).To(Equal("p@$$w0rd!#%"))
				})

				It("supports HTTPS proxy URLs", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "proxyuser",
						Password: "proxypass",
						ProxyURL: "https://secure-proxy.example.com:8443",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.ProxyAuthConfig.ProxyURL).To(Equal("https://secure-proxy.example.com:8443"))
				})

				It("supports proxy URLs with custom ports", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "proxyuser",
						Password: "proxypass",
						ProxyURL: "http://proxy.example.com:3128",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.ProxyAuthConfig.ProxyURL).To(ContainSubstring(":3128"))
				})
			})

			Context("with SPNEGO authentication", func() {
				It("returns an error when username is empty", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeSPNEGO,
						Username: "",
						Password: "password",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					_, err := tokenFetcher.GetToken()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("username"))
				})

				It("returns an error when password is empty", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeSPNEGO,
						Username: "user@REALM.COM",
						Password: "",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					_, err := tokenFetcher.GetToken()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("password"))
				})

				It("returns an error when proxy URL is empty", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeSPNEGO,
						Username: "user@REALM.COM",
						Password: "password",
						ProxyURL: "",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					_, err := tokenFetcher.GetToken()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("proxy URL is required"))
				})

				It("accepts valid SPNEGO config with Kerberos realm", func() {
					proxyAuthConfig := ProxyAuthConfig{
						AuthType:   ProxyAuthTypeSPNEGO,
						Username:   "user@REALM.COM",
						Password:   "password",
						ProxyURL:   "http://proxy.example.com:8080",
						Krb5Config: "/etc/krb5.conf",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), "some-refresh-token", false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.ProxyAuthConfig.AuthType).To(Equal(ProxyAuthTypeSPNEGO))
					Expect(tokenFetcher.ProxyAuthConfig.Krb5Config).To(Equal("/etc/krb5.conf"))
				})
			})

			Context("with different refresh tokens", func() {
				It("handles long refresh tokens with proxy auth", func() {
					longRefreshToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "proxyuser",
						Password: "proxypass",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), longRefreshToken, false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.RefreshToken).To(Equal(longRefreshToken))
				})

				It("handles short refresh tokens with proxy auth", func() {
					shortRefreshToken := "short-token-123"
					proxyAuthConfig := ProxyAuthConfig{
						AuthType: ProxyAuthTypeBasic,
						Username: "proxyuser",
						Password: "proxypass",
						ProxyURL: "http://proxy.example.com:8080",
					}
					tokenFetcher = NewTokenFetcher(server.URL(), shortRefreshToken, false, "", proxyAuthConfig)

					Expect(tokenFetcher).NotTo(BeNil())
					Expect(tokenFetcher.RefreshToken).To(Equal(shortRefreshToken))
				})
			})
		})
	})
})
