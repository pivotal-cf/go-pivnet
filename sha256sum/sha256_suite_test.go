package sha256sum_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSHA256(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SHA256 Suite")
}
