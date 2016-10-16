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

	Describe("ReleaseFingerprint", func() {
		var (
			release pivnet.Release

			releasesResponseStatusCode int
			releasesResponseBody       []byte
			releasesETagHeader         http.Header

			productFilesResponseStatusCode int
			productFilesResponseBody       []byte
			productFilesETagHeader         http.Header

			upgradePathsResponseStatusCode int
			upgradePathsResponseBody       []byte
			upgradePathsETagHeader         http.Header

			dependenciesResponseStatusCode int
			dependenciesResponseBody       []byte
			dependenciesETagHeader         http.Header
		)

		BeforeEach(func() {
			release = pivnet.Release{
				ID: 1234,
			}

			releasesResponseStatusCode = http.StatusOK
			releasesResponseBody = []byte(`{"message":"releases message"}`)
			releasesETagHeader = http.Header{"ETag": []string{`"releases-etag"`}}

			productFilesResponseStatusCode = http.StatusOK
			productFilesResponseBody = []byte(`{"message":"product files message"}`)
			productFilesETagHeader = http.Header{"ETag": []string{`"product-files-etag"`}}

			upgradePathsResponseStatusCode = http.StatusOK
			upgradePathsResponseBody = []byte(`{"message":"upgrade paths message"}`)
			upgradePathsETagHeader = http.Header{"ETag": []string{`"upgrade-paths-etag"`}}

			dependenciesResponseStatusCode = http.StatusOK
			dependenciesResponseBody = []byte(`{"message":"dependencies message"}`)
			dependenciesETagHeader = http.Header{"ETag": []string{`"dependencies-etag"`}}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(
						releasesResponseStatusCode,
						releasesResponseBody,
						releasesETagHeader,
					),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/product_files",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(
						productFilesResponseStatusCode,
						productFilesResponseBody,
						productFilesETagHeader),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/upgrade_paths",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(
						upgradePathsResponseStatusCode,
						upgradePathsResponseBody,
						upgradePathsETagHeader),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/dependencies",
						apiPrefix,
						productSlug,
						release.ID,
					)),
					ghttp.RespondWith(
						dependenciesResponseStatusCode,
						dependenciesResponseBody,
						dependenciesETagHeader),
				),
			)
		})

		It("returns the Fingerprint for the specified release", func() {
			fingerprint, err := client.ReleaseFingerprint(productSlug, release.ID)
			Expect(err).NotTo(HaveOccurred())

			// MD5 of string: "releases-etagproduct-files-etagupgrade-paths-etagdependencies-etag"
			Expect(fingerprint).To(Equal("372f8cb139e05339337895f5e1e4271c"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				releasesResponseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ReleaseFingerprint(productSlug, release.ID)
				Expect(err.Error()).To(Equal("418 - releases message. Errors: "))
			})
		})

		Context("when the etag header is missing", func() {
			BeforeEach(func() {
				releasesETagHeader = nil
			})

			It("returns an error", func() {
				_, err := client.ReleaseFingerprint(productSlug, release.ID)
				Expect(err).To(MatchError(errors.New(
					"ETag header not present")))
			})
		})

		Context("when the etag header is malformed", func() {
			var (
				malformedETag string
			)

			BeforeEach(func() {
				malformedETag = "malformed-etag-without-double-quotes"
				releasesETagHeader = http.Header{"ETag": []string{malformedETag}}
			})

			It("returns an error", func() {
				_, err := client.ReleaseFingerprint(productSlug, release.ID)
				Expect(err).To(MatchError(fmt.Errorf("ETag header malformed: %s", malformedETag)))
			})
		})

		Context("when getting the product files fails", func() {
			BeforeEach(func() {
				productFilesResponseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ReleaseFingerprint(productSlug, release.ID)
				Expect(err.Error()).To(Equal("418 - product files message. Errors: "))
			})
		})

		Context("when getting the upgrade paths fails", func() {
			BeforeEach(func() {
				upgradePathsResponseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ReleaseFingerprint(productSlug, release.ID)
				Expect(err.Error()).To(Equal("418 - upgrade paths message. Errors: "))
			})
		})

		Context("when getting the dependencies fails", func() {
			BeforeEach(func() {
				dependenciesResponseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.ReleaseFingerprint(productSlug, release.ID)
				Expect(err.Error()).To(Equal("418 - dependencies message. Errors: "))
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
