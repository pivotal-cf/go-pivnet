package main_test

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func runMainWithArgs(args ...string) *gexec.Session {
	args = append(
		args,
		fmt.Sprintf("--api-token=%s", apiToken),
		fmt.Sprintf("--endpoint=%s", endpoint),
	)

	_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", args)
	Expect(err).NotTo(HaveOccurred())

	command := exec.Command(pivnetBinPath, args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

var _ = Describe("pivnet cli", func() {
	var (
		args []string
	)

	BeforeEach(func() {
		args = []string{}
	})

	Describe("Displaying help", func() {
		It("displays help with '-h'", func() {
			session := runMainWithArgs("-h")

			Eventually(session, executableTimeout).Should(gexec.Exit())
			Expect(session.Err).Should(gbytes.Say("Usage"))
		})

		It("displays help with '--help'", func() {
			session := runMainWithArgs("--help")

			Eventually(session, executableTimeout).Should(gexec.Exit())
			Expect(session.Err).Should(gbytes.Say("Usage"))
		})
	})

	Describe("Displaying version", func() {
		It("displays version with '-v'", func() {
			session := runMainWithArgs("-v")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say("dev"))
		})

		It("displays version with '--version'", func() {
			session := runMainWithArgs("--version")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say("dev"))
		})
	})

	Describe("printing as json", func() {
		It("prints as json", func() {
			session := runMainWithArgs("--print-as=json", "product", "-s", "pivnet-resource-test")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var product pivnet.Product
			err := json.Unmarshal(session.Out.Contents(), &product)
			Expect(err).NotTo(HaveOccurred())

			Expect(product.Slug).To(Equal("pivnet-resource-test"))
		})
	})

	Describe("printing as yaml", func() {
		It("prints as yaml", func() {
			session := runMainWithArgs("--print-as=yaml", "product", "-s", "pivnet-resource-test")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var product pivnet.Product
			err := yaml.Unmarshal(session.Out.Contents(), &product)
			Expect(err).NotTo(HaveOccurred())

			Expect(product.Slug).To(Equal("pivnet-resource-test"))
		})
	})

	Describe("product", func() {
		It("displays product for the provided slug", func() {
			session := runMainWithArgs("product", "-s", "pivnet-resource-test")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say("pivnet-resource-test"))
		})
	})
})
