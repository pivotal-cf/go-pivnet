package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const (
	executableTimeout = 5 * time.Second
)

var (
	pivnetBinPath string
)

func TestCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

var _ = BeforeSuite(func() {
	By("Compiling binary")
	var err error
	pivnetBinPath, err = gexec.Build("github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet", "-race")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
