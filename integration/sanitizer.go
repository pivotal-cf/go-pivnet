package integration

import (
	"io"
	"strings"

	"github.com/onsi/ginkgo/v2"
)

type Sanitizer interface {
	ginkgo.GinkgoWriterInterface
}

type sanitizer struct {
	sanitized map[string]string
	sink      ginkgo.GinkgoWriterInterface
}

func NewSanitizer(sanitized map[string]string, sink ginkgo.GinkgoWriterInterface) Sanitizer {
	if _, ok := sanitized[""]; ok {
		delete(sanitized, "")
	}
	return &sanitizer{
		sanitized: sanitized,
		sink:      sink,
	}
}

func (s sanitizer) Print(...interface{})          {}
func (s sanitizer) Printf(string, ...interface{}) {}
func (s sanitizer) Println(...interface{})        {}
func (s sanitizer) TeeTo(io.Writer)               {}
func (s sanitizer) ClearTeeWriters()              {}

func (s sanitizer) Write(p []byte) (n int, err error) {
	input := string(p)

	for k, v := range s.sanitized {
		input = strings.Replace(input, k, v, -1)
	}

	scrubbed := []byte(input)

	return s.sink.Write(scrubbed)
}
