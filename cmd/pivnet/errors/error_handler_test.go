package errors_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errors"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer/printerfakes"
)

var _ = Describe("ErrorHandler", func() {
	var (
		fakePrinter *printerfakes.FakePrinter

		errorHandler errors.ErrorHandler
	)

	BeforeEach(func() {
		fakePrinter = &printerfakes.FakePrinter{}

		errorHandler = errors.NewErrorHandler(printer.PrintAsTable, fakePrinter)
	})

	It("returns provided error", func() {
		inputErr := fmt.Errorf("some error")

		err := errorHandler.HandleError(inputErr)
		Expect(err).To(Equal(inputErr))
	})

	It("writes to printer", func() {
		inputErr := fmt.Errorf("some error")

		_ = errorHandler.HandleError(inputErr)

		Expect(fakePrinter.PrintlnCallCount()).To(Equal(1))
	})

	Context("when the error is nil", func() {
		It("returns nil", func() {
			err := errorHandler.HandleError(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not write to printer", func() {
			_ = errorHandler.HandleError(nil)

			Expect(fakePrinter.PrintlnCallCount()).To(Equal(0))
			Expect(fakePrinter.PrintJSONCallCount()).To(Equal(0))
			Expect(fakePrinter.PrintJSONCallCount()).To(Equal(0))
		})
	})

	Describe("print as JSON", func() {
		BeforeEach(func() {
			errorHandler = errors.NewErrorHandler(printer.PrintAsJSON, fakePrinter)
		})

		It("writes to printer", func() {
			inputErr := fmt.Errorf("some error")

			_ = errorHandler.HandleError(inputErr)

			Expect(fakePrinter.PrintJSONCallCount()).To(Equal(1))
		})
	})

	Describe("print as YAML", func() {
		BeforeEach(func() {
			errorHandler = errors.NewErrorHandler(printer.PrintAsYAML, fakePrinter)
		})

		It("writes to printer", func() {
			inputErr := fmt.Errorf("some error")

			_ = errorHandler.HandleError(inputErr)

			Expect(fakePrinter.PrintYAMLCallCount()).To(Equal(1))
		})
	})
})
