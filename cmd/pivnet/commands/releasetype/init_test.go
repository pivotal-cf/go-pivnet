package releasetype_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/releasetype"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReleaseType commands suite")
}

var _ = BeforeSuite(func() {
	releasetype.Format = printer.PrintAsJSON
})
