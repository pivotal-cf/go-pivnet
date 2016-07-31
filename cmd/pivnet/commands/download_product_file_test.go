package commands_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

const (
	fileContents = "some file contents"
)

var _ = Describe("download product file commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string
		releases    []pivnet.Release

		productFile pivnet.ProductFile

		responseStatusCode int
		response           interface{}

		releasesResponseStatusCode int
		releasesResponse           pivnet.ReleasesResponse

		acceptEULAResponseStatusCode int
		acceptEULAResponse           pivnet.EULAAcceptanceResponse

		productFileResponseStatusCode int
		productFileResponse           pivnet.ProductFileResponse
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}
		commands.ErrorHandler = fakeErrorHandler

		productSlug = "some-product-slug"

		releases = []pivnet.Release{
			{
				ID:      1234,
				Version: "some-release-version",
			},
			{
				ID:      2345,
				Version: "another-release-version",
			},
		}

		productFile = pivnet.ProductFile{
			ID:   2345,
			Size: len(fileContents),
		}

		releasesResponseStatusCode = http.StatusOK
		productFileResponseStatusCode = http.StatusOK
		acceptEULAResponseStatusCode = http.StatusOK

		releasesResponse = pivnet.ReleasesResponse{
			Releases: releases,
		}

		productFileResponse = pivnet.ProductFileResponse{
			ProductFile: productFile,
		}

		acceptEULAResponse = pivnet.EULAAcceptanceResponse{}

		responseStatusCode = http.StatusOK
		response = fileContents
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("DownloadProductFileCommand", func() {
		var (
			tempDir string

			releaseVersion string
			productFileID  int
			tempFilepath   string
			acceptEULA     bool

			command commands.DownloadProductFileCommand

			origStdErr     *os.File
			testStdErrFile *os.File
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "go-pivnet")
			Expect(err).NotTo(HaveOccurred())

			releaseVersion = "some-release-version"
			productFileID = 1234
			tempFilepath = filepath.Join(tempDir, "some-file")
			acceptEULA = false

			command = commands.DownloadProductFileCommand{
				ProductSlug:    productSlug,
				ReleaseVersion: releaseVersion,
				ProductFileID:  productFileID,
				Filepath:       tempFilepath,
				AcceptEULA:     acceptEULA,
			}

			By("Replacing os.Stderr with temp file")
			testStdErrFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			origStdErr = os.Stderr
			os.Stderr = testStdErrFile
		})

		AfterEach(func() {
			os.Stderr = origStdErr

			err := os.RemoveAll(testStdErrFile.Name())
			Expect(err).NotTo(HaveOccurred())

			err = os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			command.AcceptEULA = acceptEULA

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf(
						"%s/products/%s/releases/%d/product_files/%d",
						apiPrefix,
						productSlug,
						releases[0].ID,
						productFileID,
					)),
					ghttp.RespondWithJSONEncoded(productFileResponseStatusCode, productFileResponse),
				),
			)

			if acceptEULA {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/products/%s/releases/%d/eula_acceptance",
							apiPrefix,
							productSlug,
							releases[0].ID,
						)),
						ghttp.RespondWithJSONEncoded(acceptEULAResponseStatusCode, acceptEULAResponse),
					),
				)
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf(
						"%s/products/%s/releases/%d/product_files/%d/download",
						apiPrefix,
						productSlug,
						releases[0].ID,
						productFileID,
					)),
					ghttp.RespondWith(responseStatusCode, response),
				),
			)
		})

		It("Downloads file", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			b, err := ioutil.ReadFile(tempFilepath)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(b)).To(Equal(response))
		})

		It("writes progress bar", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			b, err := ioutil.ReadFile(testStdErrFile.Name())
			Expect(err).NotTo(HaveOccurred())

			// The progress bar should look something like:
			// 18 B / 18 B [============================================] 100.00% 60.13 KB/s 0
			Expect(string(b)).To(MatchRegexp(
				`%d\ B\ /\ %d\ B\ \[=*\]\ 100.00%%`,
				len(fileContents),
				len(fileContents),
			))
		})

		Context("when there is an error before download starts", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})

			It("does not print anything to stderr (including the progress bar)", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				b, err := ioutil.ReadFile(testStdErrFile.Name())
				Expect(err).NotTo(HaveOccurred())

				Expect(b).To(BeEmpty())
			})
		})

		Context("when there is an error getting all releases", func() {
			BeforeEach(func() {
				releasesResponseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Context("when there is an error getting product file", func() {
			BeforeEach(func() {
				productFileResponseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Context("when accept-eula is true", func() {
			BeforeEach(func() {
				acceptEULA = true
			})

			It("accepts the EULA", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when accepting the EULA fails", func() {
				BeforeEach(func() {
					acceptEULAResponseStatusCode = http.StatusTeapot
				})

				It("invokes the error handler", func() {
					err := command.Execute(nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				})
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("Filepath flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "Filepath")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("filepath"))
			})
		})

		Describe("AcceptEULA flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "AcceptEULA")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("accept-eula"))
			})
		})
	})
})
