package extension_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const (
	apiPrefix   = "/api/v2"
	productSlug = "some-product-name"
)

func TestExtendedPivnetClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extended Pivnet Client Suite")
}
