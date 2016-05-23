package error

import (
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

type ErrorHandler interface {
	HandleError(err error) error
}

type errorHandler struct {
	format  string
	printer printer.Printer
}

func NewErrorHandler(format string, printer printer.Printer) ErrorHandler {
	return &errorHandler{
		format:  format,
		printer: printer,
	}
}

func (h errorHandler) HandleError(err error) error {
	var message string

	switch err.(type) {
	case pivnet.ErrUnauthorized:
		message = "Please log in first"
	case pivnet.ErrNotFound:
		message = "Not found"
	}

	switch h.format {
	case printer.PrintAsTable:
		h.printer.Println(message)
		return err
	case printer.PrintAsJSON:
		e := h.printer.PrintJSON(message)
		if e != nil {
			return e
		}
		return err
	case printer.PrintAsYAML:
		e := h.printer.PrintYAML(message)
		if e != nil {
			return e
		}
		return err
	}

	return nil
}
