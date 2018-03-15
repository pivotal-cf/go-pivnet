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

type pivnetErr struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var _ = Describe("PivnetClient", func() {
	var (
		server       *ghttp.Server
		client       pivnet.Client
		refreshToken string
		userAgent    string
		token		 string

		releases pivnet.ReleasesResponse

		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger
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
		newClientConfig = pivnet.ClientConfig{
			Host:      server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Context("when using a pivnet API token", func(){
		BeforeEach(func() {
			token = "my-auth-refreshToken"
			newClientConfig = pivnet.ClientConfig{
				Host:      server.URL(),
				Token:     token,
				UserAgent: userAgent,
			}
			client = pivnet.NewClient(newClientConfig, fakeLogger)
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

	Context("when using a UAA refreshToken", func(){
		BeforeEach(func() {
			refreshToken = "my-uaa-refreshToken-using-bearer"
			newClientConfig = pivnet.ClientConfig{
				Host:      server.URL(),
				Token:     refreshToken,
				UserAgent: userAgent,
			}
			client = pivnet.NewClient(newClientConfig, fakeLogger)
		})

		It("uses bearer authentication header", func() {
			uaaToken := "my-uaa-token"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"POST",
						fmt.Sprintf("%s/authentication/access_tokens", apiPrefix),
					),
					ghttp.VerifyJSON(fmt.Sprintf("{\"refresh_token\": \"%s\"}", refreshToken)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &pivnet.AuthResp {Token: uaaToken}),
				),
			)
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Bearer %s", uaaToken)),
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
			client = pivnet.NewClient(newClientConfig, fakeLogger)

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

		BeforeEach(func() {
			body = []byte(`{"message":"foo message"}`)
		})

		It("returns an ErrUnauthorized error with message from Pivnet", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/foo", apiPrefix),
					),
					ghttp.RespondWith(http.StatusUnauthorized, body),
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
				pivnet.ErrUnauthorized{
					ResponseCode: http.StatusUnauthorized,
					Message:      "foo message",
				},
			))
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
	})
})
