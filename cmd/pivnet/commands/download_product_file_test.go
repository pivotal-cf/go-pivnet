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
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errors/errorsfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("download product file commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorsfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string
		releases    []pivnet.Release

		responseStatusCode int
		response           interface{}

		releasesResponseStatusCode int
		releasesResponse           pivnet.ReleasesResponse
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

		fakeErrorHandler = &errorsfakes.FakeErrorHandler{}
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

		releasesResponseStatusCode = http.StatusOK

		releasesResponse = pivnet.ReleasesResponse{
			Releases: releases,
		}

		responseStatusCode = http.StatusOK
		response = "some content"
	})

	AfterEach(func() {
		server.Close()
	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
				ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
			),
		)
	})

	Describe("DownloadProductFileCommand", func() {
		var (
			tempDir string

			releaseVersion string
			productFileID  int
			tempFilepath   string

			command commands.DownloadProductFileCommand
		)

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "go-pivnet")
			Expect(err).NotTo(HaveOccurred())

			releaseVersion = "some-release-version"
			productFileID = 1234
			tempFilepath = filepath.Join(tempDir, "some-file")

			command = commands.DownloadProductFileCommand{
				ProductSlug:    productSlug,
				ReleaseVersion: releaseVersion,
				ProductFileID:  productFileID,
				Filepath:       tempFilepath,
			}
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
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

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
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
	})
})
