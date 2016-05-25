package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errors/errorsfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("release commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorsfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string

		release  pivnet.Release
		releases []pivnet.Release

		releasesResponseStatusCode int
		releasesResponse           pivnet.ReleasesResponse

		responseStatusCode int
		response           interface{}
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

		release = pivnet.Release{
			ID:      1234,
			Version: "some-release-version",
		}

		releases = []pivnet.Release{
			release,
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
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleasesCommand", func() {
		var (
			command commands.ReleasesCommand
		)

		BeforeEach(func() {
			command = commands.ReleasesCommand{
				ProductSlug: productSlug,
			}

			response = pivnet.ReleasesResponse{
				Releases: releases,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("lists all releases for the provided product slug", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedReleases []pivnet.Release

			err = json.Unmarshal(outBuffer.Bytes(), &returnedReleases)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedReleases).To(Equal(releases))
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

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleasesCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})
	})

	Describe("ReleaseCommand", func() {
		var (
			command commands.ReleaseCommand

			releaseEtagResponseStatusCode int
		)

		BeforeEach(func() {
			command = commands.ReleaseCommand{
				ProductSlug:    productSlug,
				ReleaseVersion: release.Version,
			}

			response = release

			releaseEtagResponseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
				),
			)

			// Once for the release itself
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)

			etagHeader := http.Header{
				"ETag": []string{`"some-etag"`},
			}

			// Again for the ETag
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID)),
					ghttp.RespondWithJSONEncoded(releaseEtagResponseStatusCode, response, etagHeader),
				),
			)
		})

		It("shows release for the provided product slug and release version", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedRelease commands.CLIRelease

			err = json.Unmarshal(outBuffer.Bytes(), &returnedRelease)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedRelease.ID).To(Equal(release.ID))
			Expect(returnedRelease.Version).To(Equal(release.Version))
			Expect(returnedRelease.ETag).To(Equal("some-etag"))
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

		Context("when there is an error getting the release etag", func() {
			BeforeEach(func() {
				releaseEtagResponseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("DeleteReleaseCommand", func() {
		var (
			command commands.DeleteReleaseCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			command = commands.DeleteReleaseCommand{
				ProductSlug:    productSlug,
				ReleaseVersion: release.Version,
			}

			response = release
		})

		JustBeforeEach(func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("deletes release for the provided product slug and release version", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
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
				field = fieldFor(commands.DeleteReleaseCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteReleaseCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})
})
