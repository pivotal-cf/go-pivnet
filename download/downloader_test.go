package download_test

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pivotal-cf/go-pivnet/download"
	"github.com/pivotal-cf/go-pivnet/download/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type EOFReader struct{}

func (e EOFReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

type NetError struct {
	error
}

func (ne NetError) Temporary() bool {
	return true
}

func (ne NetError) Timeout() bool {
	return true
}

var _ = Describe("Downloader", func() {
	var (
		httpClient *fakes.HTTPClient
		ranger     *fakes.Ranger
		bar        *fakes.Bar
		downloadLinkFetcher *fakes.DownloadLinkFetcher
	)

	BeforeEach(func() {
		httpClient = &fakes.HTTPClient{}
		ranger = &fakes.Ranger{}
		bar = &fakes.Bar{}

		bar.NewProxyReaderStub = func(reader io.Reader) (io.Reader) { return reader }

		downloadLinkFetcher = &fakes.DownloadLinkFetcher{}
		downloadLinkFetcher.NewDownloadLinkStub = func() (string, error) {
			return "https://example.com/some-file", nil
		}
	})

	Describe("Get", func() {
		It("writes the product to the given location", func() {
			ranger.BuildRangeReturns([]download.Range{
				{
					Lower:      0,
					Upper:      9,
					HTTPHeader: http.Header{"Range": []string{"bytes=0-9"}},
				},
				{
					Lower:      10,
					Upper:      19,
					HTTPHeader: http.Header{"Range": []string{"bytes=10-19"}},
				},
			}, nil)

			var receivedRequests []*http.Request
			var m = &sync.Mutex{}
			httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
				m.Lock()
				receivedRequests = append(receivedRequests, req)
				m.Unlock()

				switch req.Header.Get("Range") {
				case "bytes=0-9":
					return &http.Response{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("fake produ")),
					}, nil
				case "bytes=10-19":
					return &http.Response{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("ct content")),
					}, nil
				default:
					return &http.Response{
						StatusCode:    http.StatusOK,
						ContentLength: 10,
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					}, nil
				}
			}

			downloader := download.Client{
				HTTPClient: httpClient,
				Ranger:     ranger,
				Bar:        bar,
			}

			tmpFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			err = downloader.Get(tmpFile, downloadLinkFetcher, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			content, err := ioutil.ReadAll(tmpFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(Equal("fake product content"))

			Expect(ranger.BuildRangeCallCount()).To(Equal(1))
			Expect(ranger.BuildRangeArgsForCall(0)).To(Equal(int64(10)))

			Expect(bar.SetTotalArgsForCall(0)).To(Equal(int64(10)))
			Expect(bar.KickoffCallCount()).To(Equal(1))

			Expect(httpClient.DoCallCount()).To(Equal(3))

			methods := []string{
				receivedRequests[0].Method,
				receivedRequests[1].Method,
				receivedRequests[2].Method,
			}
			urls := []string{
				receivedRequests[0].URL.String(),
				receivedRequests[1].URL.String(),
				receivedRequests[2].URL.String(),
			}
			headers := []string{
				receivedRequests[1].Header.Get("Range"),
				receivedRequests[2].Header.Get("Range"),
			}

			Expect(methods).To(ConsistOf([]string{"HEAD", "GET", "GET"}))
			Expect(urls).To(ConsistOf([]string{"https://example.com/some-file", "https://example.com/some-file", "https://example.com/some-file"}))
			Expect(headers).To(ConsistOf([]string{"bytes=0-9", "bytes=10-19"}))

			Expect(bar.FinishCallCount()).To(Equal(1))
		})
	})

	Context("when a retryable error occurs", func() {
		Context("when there is an unexpected EOF", func() {
			It("successfully retries the download", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(EOFReader{}),
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}
				errors := []error{nil, nil, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{{Lower: 0, Upper: 15}}, nil)

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
				}

				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(tmpFile, downloadLinkFetcher, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				stats, err := tmpFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(stats.Size()).To(BeNumerically(">", 0))
			})
		})

		Context("when there is a temporary network error", func() {
			It("successfully retries the download", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{
						StatusCode: http.StatusPartialContent,
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}
				errors := []error{nil, NetError{errors.New("whoops")}, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{{Lower: 0, Upper: 15}}, nil)

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
				}

				tmpFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(tmpFile, downloadLinkFetcher, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				stats, err := tmpFile.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(stats.Size()).To(BeNumerically(">", 0))
			})
		})
	})

	Context("when an error occurs", func() {
		Context("when the HEAD request cannot be constucted", func() {
			It("returns an error", func() {
				downloader := download.Client{
					HTTPClient: nil,
					Ranger:     nil,
					Bar:        nil,
				}
				downloadLinkFetcher.NewDownloadLinkStub = func() (string, error) {
					return "%%%", nil
				}

				err := downloader.Get(nil, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError(ContainSubstring("failed to construct HEAD request")))
			})
		})

		Context("when the HEAD has an error", func() {
			It("returns an error", func() {
				httpClient.DoReturns(&http.Response{}, errors.New("failed request"))

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     nil,
					Bar:        nil,
				}

				err := downloader.Get(nil, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("failed to make HEAD request: failed request"))
			})
		})

		Context("when building a range fails", func() {
			It("returns an error", func() {
				httpClient.DoReturns(&http.Response{Request: &http.Request{
					URL: &url.URL{
						Scheme: "https",
						Host:   "example.com",
						Path:   "some-file",
					},
				},
				}, nil)

				ranger.BuildRangeReturns([]download.Range{}, errors.New("failed range build"))

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        nil,
				}

				err := downloader.Get(nil, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("failed to construct range: failed range build"))
			})
		})

		Context("when the GET fails", func() {
			It("returns an error", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{},
				}
				errors := []error{nil, errors.New("failed GET")}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{{}}, nil)

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
				}

				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(file, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("failed during retryable request: failed GET"))
			})
		})

		Context("when the GET returns a non-206", func() {
			It("returns an error", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("")),
					},
				}
				errors := []error{nil, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{{}}, nil)

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
				}

				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(file, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError("failed during retryable request: during GET unexpected status code was returned: 500"))
			})
		})

		Context("when the file cannot be written to", func() {
			It("returns an error", func() {
				responses := []*http.Response{
					{
						Request: &http.Request{
							URL: &url.URL{
								Scheme: "https",
								Host:   "example.com",
								Path:   "some-file",
							},
						},
					},
					{
						StatusCode: http.StatusPartialContent,
						Body:       ioutil.NopCloser(strings.NewReader("something")),
					},
				}
				errors := []error{nil, nil}

				httpClient.DoStub = func(req *http.Request) (*http.Response, error) {
					count := httpClient.DoCallCount() - 1
					return responses[count], errors[count]
				}

				ranger.BuildRangeReturns([]download.Range{{Lower: 0, Upper: 15}}, nil)

				downloader := download.Client{
					HTTPClient: httpClient,
					Ranger:     ranger,
					Bar:        bar,
				}

				closedFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())

				err = closedFile.Close()
				Expect(err).NotTo(HaveOccurred())

				err = downloader.Get(closedFile, downloadLinkFetcher, GinkgoWriter)
				Expect(err).To(MatchError(ContainSubstring("failed to write file")))
			})
		})
	})
})
