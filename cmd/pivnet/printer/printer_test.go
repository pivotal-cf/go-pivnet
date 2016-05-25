package printer_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var _ = Describe("Printer", func() {
	var (
		p printer.Printer

		w *bytes.Buffer
	)

	BeforeEach(func() {
		w = &bytes.Buffer{}

		p = printer.NewPrinter(w)
	})

	Describe("Println", func() {
		It("Prints a line", func() {
			err := p.Println("some message")

			Expect(err).NotTo(HaveOccurred())

			Expect(w.String()).To(Equal("some message\n"))
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				writer := errWriter{}
				p = printer.NewPrinter(writer)
			})

			It("returns an error", func() {
				err := p.Println("")

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("PrintJSON", func() {
		It("Prints object as JSON", func() {
			object := map[string]interface{}{
				"bar": 1234,
				"foo": "foo val",
			}
			err := p.PrintJSON(object)

			Expect(err).NotTo(HaveOccurred())

			expectedString := `{
"foo": "foo val",
"bar": 1234
}
`

			Expect(w.String()).To(MatchJSON(expectedString))
		})

		Context("when marshalling the object fails", func() {
			It("returns an error", func() {
				object := map[string]interface{}{
					"foo": make(chan string),
				}
				err := p.PrintJSON(object)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				writer := errWriter{}
				p = printer.NewPrinter(writer)
			})

			It("returns an error", func() {
				err := p.PrintJSON(struct{}{})

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("PrintYAML", func() {
		It("Prints object as YAML", func() {
			object := map[string]interface{}{
				"bar": 1234,
				"foo": "foo val",
			}
			err := p.PrintYAML(object)

			Expect(err).NotTo(HaveOccurred())

			expectedString := `---
foo: "foo val"
bar: 1234
`

			Expect(w.String()).To(MatchYAML(expectedString))
		})

		Context("when marshalling the object fails", func() {
			It("returns an error", func() {
				object := map[string]interface{}{
					"foo": make(chan string),
				}
				err := p.PrintYAML(object)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("when writing fails", func() {
			BeforeEach(func() {
				writer := errWriter{}
				p = printer.NewPrinter(writer)
			})

			It("returns an error", func() {
				err := p.PrintYAML(struct{}{})

				Expect(err).To(HaveOccurred())
			})
		})
	})
})

type errWriter struct {
}

func (w errWriter) Write([]byte) (int, error) {
	return 0, fmt.Errorf("Error writer erroring out")
}
