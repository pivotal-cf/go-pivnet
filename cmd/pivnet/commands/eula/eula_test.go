package eula_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/eula"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/eula/eulafakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var _ = Describe("eula commands", func() {
	var (
		fakePivnetClient *eulafakes.FakePivnetClient

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		outBuffer bytes.Buffer

		eulas []pivnet.EULA

		cmd *eula.EULAs
	)

	BeforeEach(func() {
		fakePivnetClient = &eulafakes.FakePivnetClient{}

		outBuffer = bytes.Buffer{}

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}

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

		fakePivnetClient.EULAsReturns(eulas, nil)
		fakePivnetClient.EULAReturns(eulas[0], nil)
		fakePivnetClient.AcceptEULAReturns(nil)

		cmd = &eula.EULAs{
			OutputWriter: &outBuffer,
			Printer:      printer.NewPrinter(&outBuffer),
			ErrorHandler: fakeErrorHandler,
			Client:       fakePivnetClient,
			Format:       printer.PrintAsJSON,
		}
	})

	Describe("EULAs", func() {
		It("lists all EULAs", func() {
			err := cmd.List(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedEULAs []pivnet.EULA
			err = json.Unmarshal(outBuffer.Bytes(), &returnedEULAs)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedEULAs).To(Equal(eulas))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("eulas error")
				fakePivnetClient.EULAsReturns(nil, expectedErr)
			})

			It("invokes the error handler", func() {
				err := cmd.List(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})

	Describe("EULACommand", func() {
		It("gets EULA", func() {
			err := cmd.Get(eulas[0].Slug)
			Expect(err).NotTo(HaveOccurred())

			var returnedEULA pivnet.EULA
			err = json.Unmarshal(outBuffer.Bytes(), &returnedEULA)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedEULA).To(Equal(eulas[0]))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("eulas error")
				fakePivnetClient.EULAReturns(pivnet.EULA{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := cmd.Get(eulas[0].Slug)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})

	Describe("AcceptEULACommand", func() {
		const (
			productSlug = "some-product-slug"
		)

		var (
			release pivnet.Release
		)

		BeforeEach(func() {
			release = pivnet.Release{
				ID:          1234,
				Version:     "version 0.2.3",
				Description: "Some release with some description.",
			}
		})

		It("accepts EULA", func() {
			err := cmd.AcceptEULA(productSlug, release.Version)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("eulas error")
				fakePivnetClient.AcceptEULAReturns(expectedErr)
			})

			It("invokes the error handler", func() {
				err := cmd.AcceptEULA(productSlug, release.Version)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})

		Context("when there is an error getting release", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("releases error")
				fakePivnetClient.ReleaseForProductVersionReturns(pivnet.Release{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := cmd.AcceptEULA(productSlug, release.Version)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})
})
