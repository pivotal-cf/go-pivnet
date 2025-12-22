package pivnet_test

import (
	"fmt"
	"net/http"

	gopivnetfakes "github.com/pivotal-cf/go-pivnet/v9/go-pivnetfakes"

	"github.com/onsi/gomega/ghttp"

	"github.com/pivotal-cf/go-pivnet/v9"
	"github.com/pivotal-cf/go-pivnet/v9/logger"
	"github.com/pivotal-cf/go-pivnet/v9/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type pivnetErr struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var _ = Describe("PivnetClient", func() {
	var (
		server    *ghttp.Server
		client    pivnet.Client
		userAgent string
		token     string

		releases pivnet.ReleasesResponse

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
	)

	BeforeEach(func() {
		releases = pivnet.ReleasesResponse{Releases: []pivnet.Release{
			{
				ID:      1,
				Version: "1234",
			},
			{
				ID:      99,
				Version: "some-other-version",
			},
		}}

		server = ghttp.NewServer()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host:      server.URL(),
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
	})

	JustBeforeEach(func() {
		fakeAccessTokenService.AccessTokenReturns(token, nil)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("when using a pivnet API token", func() {
		BeforeEach(func() {
			newClientConfig = pivnet.ClientConfig{
				Host:      server.URL(),
				UserAgent: userAgent,
			}
			client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
		})
		It("uses token authentication header if configured with a pivnet api token", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when using a UAA refreshToken", func() {
		BeforeEach(func() {
			token = "my-auth-token-that-is-longer-than-a-legacy-token"
			newClientConfig = pivnet.ClientConfig{
				Host:      server.URL(),
				UserAgent: userAgent,
			}
			client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
		})

		It("uses bearer authentication header", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Bearer %s", token)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("sets custom user agent", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/foo", apiPrefix),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.VerifyHeaderKV("User-Agent", userAgent),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		_, err := client.MakeRequest(
			"GET",
			"/foo",
			http.StatusOK,
			nil,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets Content-Type application/json", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/foo", apiPrefix),
				),
				ghttp.VerifyHeaderKV("Content-Type", "application/json"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		_, err := client.MakeRequest(
			"GET",
			"/foo",
			http.StatusOK,
			nil,
		)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when parsing the url fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Host = "%%%"
			client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("%%%"))
		})
	})

	Context("when Pivnet returns a 401", func() {
		var (
			body []byte
		)

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusUnauthorized, body),
				),
			)
		})

		Context("when Pivnet returns JSON", func() {
			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an ErrUnauthorized error with message from Pivnet", func() {
				_, err := client.MakeRequest(
					"GET",
					"/foo",
					http.StatusOK,
					nil,
				)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(
					pivnet.ErrUnauthorized{
						ResponseCode: http.StatusUnauthorized,
						Message:      "foo message",
					},
				))
			})
		})

		Context("when Pivnet returns a non JSON", func() {
			BeforeEach(func() {
				body = []byte("Forbidden")
			})

			It("returns a well formatted error", func() {
				_, err := client.MakeRequest(
					"GET",
					"/foo",
					http.StatusOK,
					nil,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse json [\"Forbidden\"]"))
			})
		})
	})

	Context("when Pivnet returns a 429", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`Retry later`)
		})

		It("returns an ErrUnauthorized error with message from Pivnet", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusTooManyRequests, body),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				pivnet.ErrTooManyRequests{
					ResponseCode: http.StatusTooManyRequests,
					Message:      "You have hit a rate limit for this request",
				},
			))
		})
	})

	Context("when Pivnet returns a 451", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"message":"I should be visible to the user"}`)
		})

		It("returns an ErrUnavailableForLegalReasons error with message from Pivnet", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusUnavailableForLegalReasons, body),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				pivnet.ErrUnavailableForLegalReasons{
					ResponseCode: http.StatusUnavailableForLegalReasons,
					Message:      "I should be visible to the user",
				},
			))
		})
	})

	Context("when Pivnet returns a 404", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"message":"foo message"}`)
		})

		It("returns an ErrNotFound error with message from Pivnet", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusNotFound, body),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				pivnet.ErrNotFound{
					ResponseCode: http.StatusNotFound,
					Message:      "foo message",
				},
			))
		})
	})

	Context("when Pivnet returns a 500", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"status":"500","error":"foo message"}`)
		})

		It("returns an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusInternalServerError, body),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				pivnet.ErrPivnetOther{
					ResponseCode: http.StatusInternalServerError,
					Message:      "foo message",
				},
			))
		})

		Context("when unmarshalling the response from Pivnet returns an error", func() {
			BeforeEach(func() {
				body = []byte(`{"error":1234}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/foo", apiPrefix),
						),
						ghttp.RespondWith(http.StatusInternalServerError, body),
					),
				)

				_, err := client.MakeRequest(
					"GET",
					"/foo",
					http.StatusOK,
					nil,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("json: cannot unmarshal"))
			})
		})
	})

	Context("when an unexpected status code comes back from Pivnet", func() {
		var (
			body []byte
		)

		BeforeEach(func() {
			body = []byte(`{"message":"foo message"}`)
		})

		It("returns an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusTeapot, body),
				),
			)

			_, err := client.MakeRequest(
				"GET",
				"/foo",
				http.StatusOK,
				nil,
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("foo message"))
		})

		Context("when unmarshalling the response from Pivnet returns an error", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/foo", apiPrefix),
						),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.MakeRequest(
					"GET",
					"/foo",
					http.StatusOK,
					nil,
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("JSON"))
			})
		})

		Context("when an expectedResponseCode of 0 is provided", func() {
			It("does not return an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/foo", apiPrefix),
						),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.MakeRequest(
					"GET",
					"/foo",
					0,
					nil,
				)
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})

	Describe("CreateRequest", func() {
		It("strips the host prefix if present", func() {
			req, err := client.CreateRequest(
				"GET",
				fmt.Sprintf("https://example.com/%s/foo/bar", "api/v2"),
				nil,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(req.URL.Path).To(Equal("/api/v2/foo/bar"))
		})

		It("does not add auth header if versions endpoint", func() {
			req, err := client.CreateRequest(
				"GET",
				"/versions",
				nil,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(req.Header.Get("Authorization")).To(Equal(""))
			Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))
		})
	})

	Describe("NewClientWithProxy", func() {
		var (
			proxyConfig pivnet.ClientConfig
		)

		BeforeEach(func() {
			proxyConfig = pivnet.ClientConfig{
				Host:      server.URL(),
				UserAgent: userAgent,
			}
		})

		Context("when no proxy auth is configured", func() {
			It("creates a client successfully", func() {
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("creates a client with default transport", func() {
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.HTTP).NotTo(BeNil())
			})

			It("initializes all client services", func() {
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Auth).NotTo(BeNil())
				Expect(client.EULA).NotTo(BeNil())
				Expect(client.ProductFiles).NotTo(BeNil())
				Expect(client.FileGroups).NotTo(BeNil())
				Expect(client.Releases).NotTo(BeNil())
				Expect(client.Products).NotTo(BeNil())
				Expect(client.UserGroups).NotTo(BeNil())
			})
		})

		Context("when Basic proxy auth is configured", func() {
			BeforeEach(func() {
				proxyConfig.ProxyAuthConfig = pivnet.ProxyAuthConfig{
					AuthType: pivnet.ProxyAuthTypeBasic,
					ProxyURL: "http://proxy.example.com:8080",
					Username: "proxyuser",
					Password: "proxypass",
				}
			})

			It("creates a client with proxy auth transport", func() {
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("creates a client with custom transport", func() {
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.HTTP).NotTo(BeNil())
				Expect(client.HTTP.Transport).NotTo(BeNil())
			})

			It("accepts special characters in username", func() {
				proxyConfig.ProxyAuthConfig.Username = "user@domain.com"
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("accepts special characters in password", func() {
				proxyConfig.ProxyAuthConfig.Password = "p@$$w0rd!#%"
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("accepts HTTPS proxy URL", func() {
				proxyConfig.ProxyAuthConfig.ProxyURL = "https://proxy.example.com:8443"
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})

		Context("when SPNEGO proxy auth is configured", func() {
			BeforeEach(func() {
				proxyConfig.ProxyAuthConfig = pivnet.ProxyAuthConfig{
					AuthType:   pivnet.ProxyAuthTypeSPNEGO,
					ProxyURL:   "http://proxy.example.com:8080",
					Username:   "user@REALM.COM",
					Password:   "password",
					Krb5Config: "/tmp/krb5.conf",
				}
			})

			It("returns an error when Kerberos login fails", func() {
				// SPNEGO will fail because there's no real KDC
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create proxy authenticator"))
			})

			It("returns an error with empty username", func() {
				proxyConfig.ProxyAuthConfig.Username = ""
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
			})

			It("returns an error with empty password", func() {
				proxyConfig.ProxyAuthConfig.Password = ""
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
			})

			It("returns an error with empty Krb5Config", func() {
				proxyConfig.ProxyAuthConfig.Krb5Config = ""
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when proxy auth type is invalid", func() {
			BeforeEach(func() {
				proxyConfig.ProxyAuthConfig = pivnet.ProxyAuthConfig{
					AuthType: "invalid",
					ProxyURL: "http://proxy.example.com:8080",
				}
			})

			It("returns an error", func() {
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create proxy authenticator"))
			})
		})

		Context("when proxy URL is empty but auth type is set", func() {
			BeforeEach(func() {
				proxyConfig.ProxyAuthConfig = pivnet.ProxyAuthConfig{
					AuthType: pivnet.ProxyAuthTypeBasic,
					ProxyURL: "",
					Username: "user",
					Password: "pass",
				}
			})

			It("returns an error", func() {
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("proxy URL is required"))
			})
		})

		Context("when creating proxy auth transport fails", func() {
			BeforeEach(func() {
				proxyConfig.ProxyAuthConfig = pivnet.ProxyAuthConfig{
					AuthType: pivnet.ProxyAuthTypeBasic,
					ProxyURL: "", // Empty URL will cause validation error
				}
			})

			It("returns an error", func() {
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("proxy URL is required"))
			})
		})

		Context("with SkipSSLValidation", func() {
			BeforeEach(func() {
				proxyConfig.SkipSSLValidation = true
			})

			It("creates a client with SSL validation skipped", func() {
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("works with Basic proxy auth", func() {
				proxyConfig.ProxyAuthConfig = pivnet.ProxyAuthConfig{
					AuthType: pivnet.ProxyAuthTypeBasic,
					ProxyURL: "http://proxy.example.com:8080",
					Username: "user",
					Password: "pass",
				}
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, proxyConfig, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})
		})
	})

	Describe("ProxyAuthConfig", func() {
		Context("validation", func() {
			It("accepts empty ProxyAuthConfig", func() {
				config := pivnet.ClientConfig{
					Host:            server.URL(),
					UserAgent:       userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{},
				}
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("accepts Basic auth with empty username", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: pivnet.ProxyAuthTypeBasic,
						ProxyURL: "http://proxy.example.com:8080",
						Username: "",
						Password: "password",
					},
				}
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("accepts Basic auth with empty password", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: pivnet.ProxyAuthTypeBasic,
						ProxyURL: "http://proxy.example.com:8080",
						Username: "username",
						Password: "",
					},
				}
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("accepts Basic auth with both username and password empty", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: pivnet.ProxyAuthTypeBasic,
						ProxyURL: "http://proxy.example.com:8080",
						Username: "",
						Password: "",
					},
				}
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("rejects SPNEGO auth without username", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType:   pivnet.ProxyAuthTypeSPNEGO,
						ProxyURL:   "http://proxy.example.com:8080",
						Username:   "",
						Password:   "password",
						Krb5Config: "/tmp/krb5.conf",
					},
				}
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).To(HaveOccurred())
			})

			It("rejects SPNEGO auth without password", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType:   pivnet.ProxyAuthTypeSPNEGO,
						ProxyURL:   "http://proxy.example.com:8080",
						Username:   "user@REALM.COM",
						Password:   "",
						Krb5Config: "/tmp/krb5.conf",
					},
				}
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).To(HaveOccurred())
			})

			It("rejects SPNEGO auth without Krb5Config", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType:   pivnet.ProxyAuthTypeSPNEGO,
						ProxyURL:   "http://proxy.example.com:8080",
						Username:   "user@REALM.COM",
						Password:   "password",
						Krb5Config: "",
					},
				}
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).To(HaveOccurred())
			})

			It("rejects proxy auth without proxy URL", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: pivnet.ProxyAuthTypeBasic,
						ProxyURL: "",
						Username: "user",
						Password: "pass",
					},
				}
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("proxy URL is required"))
			})
		})

		Context("auth type constants", func() {
			It("accepts ProxyAuthTypeBasic constant", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: pivnet.ProxyAuthTypeBasic,
						ProxyURL: "http://proxy.example.com:8080",
						Username: "user",
						Password: "pass",
					},
				}
				client, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
			})

			It("rejects uppercase 'BASIC' auth type", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: "BASIC",
						ProxyURL: "http://proxy.example.com:8080",
						Username: "user",
						Password: "pass",
					},
				}
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported proxy authentication type"))
			})

			It("rejects mixed case 'Basic' auth type", func() {
				config := pivnet.ClientConfig{
					Host:      server.URL(),
					UserAgent: userAgent,
					ProxyAuthConfig: pivnet.ProxyAuthConfig{
						AuthType: "Basic",
						ProxyURL: "http://proxy.example.com:8080",
						Username: "user",
						Password: "pass",
					},
				}
				_, err := pivnet.NewClientWithProxy(fakeAccessTokenService, config, fakeLogger)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported proxy authentication type"))
			})
		})
	})
})
