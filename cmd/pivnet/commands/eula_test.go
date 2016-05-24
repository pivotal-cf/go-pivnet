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

var _ = Describe("eula commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorsfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		eulas []pivnet.EULA

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

		eulas = []pivnet.EULA{
			{
				ID:   1234,
				Name: "some eula",
				Slug: "some-eula",
			},
			{
				ID:   2345,
				Name: "another eula",
				Slug: "another-eula",
			},
		}

		responseStatusCode = http.StatusOK
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("EULAsCommand", func() {
		var (
			command commands.EULAsCommand
		)

		BeforeEach(func() {
			response = pivnet.EULAsResponse{
				EULAs: eulas,
			}

			command = commands.EULAsCommand{}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("lists all EULAs", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedEULAs []pivnet.EULA

			err = json.Unmarshal(outBuffer.Bytes(), &returnedEULAs)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedEULAs).To(Equal(eulas))
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
	})

	Describe("EULACommand", func() {
		var (
			command commands.EULACommand
		)

		BeforeEach(func() {
			response = eulas[0]

			command = commands.EULACommand{
				EULASlug: eulas[0].Slug,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas/%s", apiPrefix, eulas[0].Slug)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("shows EULA", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedEULA pivnet.EULA

			err = json.Unmarshal(outBuffer.Bytes(), &returnedEULA)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedEULA).To(Equal(eulas[0]))
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

		Describe("EULASlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.EULACommand{}, "EULASlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("eula-slug"))
			})
		})
	})

	Describe("AcceptEULACommand", func() {
		const (
			productSlug = "some-product-slug"
		)

		var (
			releases []pivnet.Release
			command  commands.AcceptEULACommand
		)

		BeforeEach(func() {
			response = pivnet.EULAAcceptanceResponse{
				AcceptedAt: "now",
			}

			releases = []pivnet.Release{
				{
					ID:          1234,
					Version:     "version 0.2.3",
					Description: "Some release with some description.",
				},
				{
					ID:          2345,
					Version:     "version 0.3.4",
					Description: "Another release with another description.",
				},
			}

			command = commands.AcceptEULACommand{
				ProductSlug:    productSlug,
				ReleaseVersion: releases[0].Version,
			}
		})

		JustBeforeEach(func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"POST",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/eula_acceptance",
							apiPrefix,
							productSlug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("accepts EULA", func() {
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

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AcceptEULACommand{}, "ProductSlug")
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
				field = fieldFor(commands.AcceptEULACommand{}, "ReleaseVersion")
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
