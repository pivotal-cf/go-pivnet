package main_test

import (
	"os"
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
	apiToken      string
	endpoint      string
)

func TestCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

var _ = BeforeSuite(func() {
	apiToken = os.Getenv("API_TOKEN")
	endpoint = os.Getenv("ENDPOINT")

	if apiToken == "" {
		Fail("API_TOKEN must be set for CLI tests to run")
	}

	if endpoint == "" {
		Fail("ENDPOINT must be set for CLI tests to run")
	}

	By("Compiling binary")
	var err error
	pivnetBinPath, err = gexec.Build("github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet", "-race")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
