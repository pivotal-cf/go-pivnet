package integration_test

import (
	"fmt"
	"os"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/robdimsdale/sanitizer"

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

	By("Sanitizing acceptance test output")
	sanitized := map[string]string{
		APIToken: "***sanitized-api-token***",
	}
	sanitizedWriter := sanitizer.NewSanitizer(sanitized, GinkgoWriter)
	GinkgoWriter = sanitizedWriter

	config := pivnet.ClientConfig{
		Host:      Host,
		Token:     APIToken,
		UserAgent: "go-pivnet/integration-test",
	}

	logger := GinkgoLogShim{}

	client = pivnet.NewClient(config, logger)

	ok, err := client.Auth.Check()
	Expect(err).NotTo(HaveOccurred())
	Expect(ok).To(BeTrue())
})

type GinkgoLogShim struct {
}

func (l GinkgoLogShim) Debug(action string, data ...logger.Data) {
	l.Info(action, data...)
}

func (l GinkgoLogShim) Info(action string, data ...logger.Data) {
	GinkgoWriter.Write([]byte(fmt.Sprintf("%s%s\n", action, appendString(data...))))
}

func appendString(data ...logger.Data) string {
	if len(data) > 0 {
		return fmt.Sprintf(" - %+v", data)
	}
	return ""
}
