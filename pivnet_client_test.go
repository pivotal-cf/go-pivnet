package pivnet_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("PivnetClient", func() {
	var (
		server    *ghttp.Server
		client    pivnet.Client
		token     string
		userAgent string

		releases   pivnet.ReleasesResponse
		etagHeader []http.Header

		newClientConfig pivnet.ClientConfig
		fakeLogger      lager.Logger
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

		etagHeader = []http.Header{
			{"ETag": []string{`"etag-0"`}},
			{"ETag": []string{`"etag-1"`}},
		}

		server = ghttp.NewServer()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = lager.NewLogger("pivnet client")
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

	It("has authenticated headers for each request", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		for i, r := range releases.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.RespondWith(http.StatusOK, nil, etagHeader[i]),
				),
			)
		}

		_, err := client.Releases.GetByProductSlug(productSlug)
		Expect(err).NotTo(HaveOccurred())
	})

	It("sets custom user agent", func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"GET",
					fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
				),
				ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
				ghttp.VerifyHeaderKV("User-Agent", userAgent),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releases),
			),
		)

		for i, r := range releases.Releases {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, r.ID),
					),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", token)),
					ghttp.VerifyHeaderKV("User-Agent", userAgent),
					ghttp.RespondWith(http.StatusOK, nil, etagHeader[i]),
				),
			)
		}

		_, err := client.Releases.GetByProductSlug(productSlug)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when parsing the url fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Host = "%%%"
			client = pivnet.NewClient(newClientConfig, fakeLogger)

			_, err := client.Releases.GetByProductSlug("some product")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("%%%"))
		})
	})

	Context("when making the request fails with error", func() {
		It("forwards the error", func() {
			newClientConfig.Host = "https://not-a-real-url.com"
			client = pivnet.NewClient(newClientConfig, fakeLogger)

			_, err := client.Releases.GetByProductSlug("some-product")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when a non-200 comes back from Pivnet", func() {
		It("returns an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusNotFound, nil),
				),
			)

			_, err := client.Releases.GetByProductSlug("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(
				"Pivnet returned status code: 404 for the request - expected 200"))
		})
	})

	Context("when the json unmarshalling fails with error", func() {
		It("forwards the error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/my-product-id/releases"),
					ghttp.RespondWith(http.StatusOK, "%%%"),
				),
			)

			_, err := client.Releases.GetByProductSlug("my-product-id")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})
})
