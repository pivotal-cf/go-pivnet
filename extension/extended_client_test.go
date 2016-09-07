package extension_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/extension"
	"github.com/pivotal-cf/go-pivnet/extension/extensionfakes"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
)

var _ = Describe("ExtendedClient", func() {
	var (
		server     *ghttp.Server
		fakeLogger *loggerfakes.FakeLogger
		client     extension.ExtendedClient
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		token := "my-auth-token"
		userAgent := "go-pivnet/0.1.0"

		fakeLogger = &loggerfakes.FakeLogger{}
		newClientConfig := pivnet.ClientConfig{
			Host:      server.URL(),
			Token:     token,
			UserAgent: userAgent,
		}
		c := pivnet.NewClient(newClientConfig, fakeLogger)
		client = extension.NewExtendedClient(c, fakeLogger)
	})

	Describe("ReleaseETag", func() {
		var (
			release pivnet.Release
		)

		BeforeEach(func() {
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
				Expect(err.Error()).To(Equal("418 - foo message. Errors: "))
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

	Describe("DownloadFile", func() {
		var (
			downloadLink string

			fileContents []byte

			httpStatus int
		)

		BeforeEach(func() {
			downloadLink = "/some/download/link"

			fileContents = []byte("some file contents")

			httpStatus = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s%s",
						apiPrefix,
						downloadLink,
					)),
					ghttp.RespondWith(httpStatus, fileContents),
				),
			)
		})

		It("writes file contents to provided writer", func() {
			writer := bytes.NewBuffer(nil)

			err := client.DownloadFile(writer, downloadLink)
			Expect(err).NotTo(HaveOccurred())

			Expect(writer.Bytes()).To(Equal(fileContents))
		})

		Context("when creating the request returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("some request error")
				fakeC := &extensionfakes.FakeClient{}
				fakeC.CreateRequestReturns(nil, expectedErr)

				client = extension.NewExtendedClient(fakeC, fakeLogger)
			})

			It("forwards the error", func() {
				err := client.DownloadFile(nil, downloadLink)
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("when dumping the request returns an error", func() {
			BeforeEach(func() {
				u, err := url.Parse("https://example.com")
				Expect(err).NotTo(HaveOccurred())

				request := &http.Request{
					URL: u,
				}

				fakeC := &extensionfakes.FakeClient{}
				fakeC.CreateRequestReturns(request, nil)

				client = extension.NewExtendedClient(fakeC, fakeLogger)
			})

			It("forwards the error", func() {
				err := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when making the request returns an error", func() {
			BeforeEach(func() {
				u, err := url.Parse("https://not-a-real-site-5463456.com")
				Expect(err).NotTo(HaveOccurred())

				request := &http.Request{
					Header: http.Header{},
					URL:    u,
				}

				fakeC := &extensionfakes.FakeClient{}
				fakeC.CreateRequestReturns(request, nil)

				client = extension.NewExtendedClient(fakeC, fakeLogger)
			})

			It("forwards the error", func() {
				err := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the response status code is 451", func() {
			BeforeEach(func() {
				httpStatus = http.StatusUnavailableForLegalReasons
			})

			It("returns an error", func() {
				err := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("EULA"))
			})
		})

		Context("when the response status code is not 200", func() {
			BeforeEach(func() {
				httpStatus = http.StatusTeapot
			})

			It("returns an error", func() {
				err := client.DownloadFile(nil, downloadLink)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("418"))
			})
		})

		Context("when there is an error copying the contents", func() {
			var (
				writer errWriter
			)

			BeforeEach(func() {
				writer = errWriter{}
			})

			It("returns an error", func() {
				err := client.DownloadFile(writer, downloadLink)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("error writing"))
			})
		})
	})
})

type errWriter struct {
}

func (e errWriter) Write([]byte) (int, error) {
	return 0, errors.New("error writing")
}
