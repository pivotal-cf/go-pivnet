package releasetype_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/releasetype"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/releasetype/releasetypefakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var _ = Describe("releasetype commands", func() {
	var (
		fakePivnetClient *releasetypefakes.FakePivnetClient

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		outBuffer bytes.Buffer

		releasetypes []string
	)

	BeforeEach(func() {
		fakePivnetClient = &releasetypefakes.FakePivnetClient{}

		outBuffer = bytes.Buffer{}
		releasetype.OutputWriter = &outBuffer
		releasetype.Printer = printer.NewPrinter(&outBuffer)

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}
		releasetype.ErrorHandler = fakeErrorHandler

		releasetypes = []string{
			"release-type-A",
			"release-type-B",
		}

		fakePivnetClient.ReleaseTypesReturns(releasetypes, nil)
		releasetype.Client = fakePivnetClient
	})

	Describe("ReleaseTypesCommand", func() {
		var (
			cmd releasetype.ReleaseTypesCommand
		)

		BeforeEach(func() {
			cmd = releasetype.ReleaseTypesCommand{}
		})

		It("lists all ReleaseTypes", func() {
			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedReleaseTypes []string
			err = json.Unmarshal(outBuffer.Bytes(), &returnedReleaseTypes)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedReleaseTypes).To(Equal(releasetypes))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("releasetypes error")
				fakePivnetClient.ReleaseTypesReturns(nil, expectedErr)
			})

			It("invokes the error handler", func() {
				err := cmd.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})
})
