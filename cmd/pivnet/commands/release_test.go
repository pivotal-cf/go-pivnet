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
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("release commands", func() {
	var (
		server *ghttp.Server

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string

		release  pivnet.Release
		releases []pivnet.Release
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

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
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleasesCommand", func() {
		It("lists all releases for the provided product slug", func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			releasesCommand := commands.ReleasesCommand{
				ProductSlug: productSlug,
			}

			err := releasesCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedReleases []pivnet.Release

			err = json.Unmarshal(outBuffer.Bytes(), &returnedReleases)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedReleases).To(Equal(releases))
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
		It("shows release for the provided product slug and release version", func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			releaseResponse := release

			// Once for the release itself
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseResponse),
				),
			)

			etagHeader := http.Header{
				"ETag": []string{`"some-etag"`},
			}

			// Again for the ETag
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseResponse, etagHeader),
				),
			)

			releaseCommand := commands.ReleaseCommand{
				ProductSlug:    productSlug,
				ReleaseVersion: release.Version,
			}

			err := releaseCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedRelease commands.CLIRelease

			err = json.Unmarshal(outBuffer.Bytes(), &returnedRelease)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedRelease.ID).To(Equal(release.ID))
			Expect(returnedRelease.Version).To(Equal(release.Version))
			Expect(returnedRelease.ETag).To(Equal("some-etag"))
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
		It("deletes release for the provided product slug and release version", func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			releaseResponse := release

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, productSlug, release.ID)),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, releaseResponse),
				),
			)

			deleteReleaseCommand := commands.DeleteReleaseCommand{
				ProductSlug:    productSlug,
				ReleaseVersion: release.Version,
			}

			err := deleteReleaseCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
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
