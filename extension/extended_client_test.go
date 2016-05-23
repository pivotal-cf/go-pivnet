package extension_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
	"github.com/pivotal-cf-experimental/go-pivnet/logger/loggerfakes"
)

var _ = Describe("ReleaseETag", func() {
	var (
		server  *ghttp.Server
		client  extension.ExtendedClient
		release pivnet.Release
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		fakeLogger := &loggerfakes.FakeLogger{}
		token := "my-auth-token"
		userAgent := "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		newClientConfig := pivnet.ClientConfig{
			Host:      server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}
		c := pivnet.NewClient(newClientConfig, fakeLogger)
		client = extension.NewExtendedClient(c, fakeLogger)

		release = pivnet.Release{
			ID: 1234,
		}
	})

	It("returns the ETag for the specified release", func() {
		etagHeader := http.Header{"ETag": []string{`"etag-0"`}}

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf(
					"%s/products/%s/releases/%d",
					apiPrefix,
					productSlug,
					release.ID,
				)),
				ghttp.RespondWith(http.StatusOK, nil, etagHeader),
			),
		)

		etag, err := client.ReleaseETag(productSlug, release.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(etag).To(Equal("etag-0"))
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
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(http.StatusTeapot, body),
				),
			)

			_, err := client.ReleaseETag(productSlug, release.ID)
			Expect(err).To(MatchError(errors.New(
				"Pivnet returned status code: 418 for the request - expected 200")))
		})
	})

	Context("when the etag is missing", func() {
		It("returns an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)

			_, err := client.ReleaseETag(productSlug, release.ID)
			Expect(err).To(MatchError(errors.New(
				"ETag header not present")))
		})
	})

	Context("when the etag is malformed", func() {
		It("returns an error", func() {
			malformedETag := "malformed-etag-without-double-quotes"
			etagHeader := http.Header{"ETag": []string{malformedETag}}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(http.StatusOK, nil, etagHeader),
				),
			)

			_, err := client.ReleaseETag(productSlug, release.ID)
			Expect(err).To(MatchError(fmt.Errorf("ETag header malformed: %s", malformedETag)))
		})
	})
})
