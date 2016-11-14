package integration_test

import (
	"os"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

const testProductSlug = "pivnet-resource-test"

var client pivnet.Client

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	APIToken := os.Getenv("API_TOKEN")
	Host := os.Getenv("HOST")

	if APIToken == "" {
		Fail("API_TOKEN must be set for integration tests to run")
	}

	if Host == "" {
		Fail("HOST must be set for integration tests to run")
	}

	config := pivnet.ClientConfig{
		Host:      Host,
		Token:     APIToken,
		UserAgent: "go-pivnet/integration-test",
	}

	logger := &loggerfakes.FakeLogger{}

	client = pivnet.NewClient(config, logger)

	ok, err := client.Auth.Check()
	Expect(err).NotTo(HaveOccurred())
	Expect(ok).To(BeTrue())
})
