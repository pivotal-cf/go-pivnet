package main_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

const (
	apiPrefix = "/api/v2"
	apiToken  = "some-api-token"
)

var _ = Describe("pivnet cli", func() {
	var (
		server *ghttp.Server
		host   string

		product pivnet.Product

		releases []pivnet.Release
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		host = server.URL()

		product = pivnet.Product{
			ID:   1234,
			Slug: "some-product-slug",
			Name: "some-product-name",
		}

		releases = []pivnet.Release{
			{
				ID:          1234,
				Version:     "version 0.2.3",
				Description: "Some release with some description.",
			},
			{
				ID:          2345,
				Version:     "version 0.3.4",
				Description: "Another release with another description.",
			},
		}

	})

	runMainWithArgs := func(args ...string) *gexec.Session {
		args = append(
			args,
			fmt.Sprintf("--api-token=%s", apiToken),
			fmt.Sprintf("--host=%s", host),
		)

		_, err := fmt.Fprintf(GinkgoWriter, "Running command: %v\n", args)
		Expect(err).NotTo(HaveOccurred())

		command := exec.Command(pivnetBinPath, args...)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		return session
	}

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
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, product),
				),
			)
		})

		It("prints as json", func() {
			session := runMainWithArgs(
				"--format=json",
				"product",
				"--product-slug", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var receivedProduct pivnet.Product
			err := json.Unmarshal(session.Out.Contents(), &receivedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedProduct.Slug).To(Equal(product.Slug))
		})
	})

	Describe("printing as yaml", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, product),
				),
			)
		})

		It("prints as yaml", func() {
			session := runMainWithArgs(
				"--format=yaml",
				"product",
				"--product-slug", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var receivedProduct pivnet.Product
			err := yaml.Unmarshal(session.Out.Contents(), &receivedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedProduct.Slug).To(Equal(product.Slug))
		})
	})

	Describe("product-files", func() {
		var (
			productFiles []pivnet.ProductFile
		)

		BeforeEach(func() {
			productFiles = []pivnet.ProductFile{
				pivnet.ProductFile{
					ID:   1234,
					Name: "some-product-file",
				},
				pivnet.ProductFile{
					ID:   2345,
					Name: "some-other-product-file",
				},
			}

		})

		It("displays product files", func() {
			response := pivnet.ProductFilesResponse{
				ProductFiles: productFiles,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/product_files",
							apiPrefix,
							product.Slug,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			session := runMainWithArgs(
				"product-files",
				"--product-slug", product.Slug,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(productFiles[0].Name))
			Expect(session).Should(gbytes.Say(productFiles[1].Name))
		})

		Context("when release version is provided", func() {
			BeforeEach(func() {
				releasesResponse := pivnet.ReleasesResponse{
					Releases: releases,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/products/%s/releases", apiPrefix, product.Slug),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
					),
				)

				response := pivnet.ProductFilesResponse{
					ProductFiles: productFiles,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/products/%s/releases/%d/product_files",
								apiPrefix,
								product.Slug,
								releases[0].ID,
							),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, response),
					),
				)
			})

			It("displays product files for release", func() {
				session := runMainWithArgs(
					"product-files",
					"--product-slug", product.Slug,
					"--release-version", releases[0].Version,
				)

				Eventually(session, executableTimeout).Should(gexec.Exit(0))
				Expect(session).Should(gbytes.Say(productFiles[0].Name))
				Expect(session).Should(gbytes.Say(productFiles[1].Name))
			})
		})
	})

	Describe("product-file", func() {
		var (
			productFile pivnet.ProductFile
		)

		BeforeEach(func() {
			productFile = pivnet.ProductFile{
				ID:   1234,
				Name: "some-product-file",
			}
		})

		It("displays product file", func() {
			response := pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/product_files/%d",
							apiPrefix,
							product.Slug,
							productFile.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			session := runMainWithArgs(
				"product-file",
				"--product-slug", product.Slug,
				"--product-file-id", strconv.Itoa(productFile.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(productFile.Name))
		})

		Context("when release version is provided", func() {
			BeforeEach(func() {
				releasesResponse := pivnet.ReleasesResponse{
					Releases: releases,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/products/%s/releases", apiPrefix, product.Slug),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
					),
				)
			})

			It("displays product files for release", func() {
				response := pivnet.ProductFileResponse{
					ProductFile: productFile,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf("%s/products/%s/releases/%d/product_files/%d",
								apiPrefix,
								product.Slug,
								releases[0].ID,
								productFile.ID,
							),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, response),
					),
				)

				session := runMainWithArgs(
					"product-file",
					"--product-slug", product.Slug,
					"--release-version", releases[0].Version,
					"--product-file-id", strconv.Itoa(productFile.ID),
				)

				Eventually(session, executableTimeout).Should(gexec.Exit(0))
				Expect(session).Should(gbytes.Say(productFile.Name))
			})
		})
	})

	Describe("add product-file", func() {
		var (
			productFile pivnet.ProductFile
		)

		BeforeEach(func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			productFile = pivnet.ProductFile{
				ID:   1234,
				Name: "some-product-file",
			}

			response := pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/add_product_file",
							apiPrefix,
							product.Slug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, response),
				),
			)
		})

		It("adds product file", func() {
			session := runMainWithArgs(
				"add-product-file",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
				"--product-file-id", strconv.Itoa(productFile.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
		})
	})

	Describe("remove product-file", func() {
		var (
			productFile pivnet.ProductFile
		)

		BeforeEach(func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases", apiPrefix, product.Slug),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			productFile = pivnet.ProductFile{
				ID:   1234,
				Name: "some-product-file",
			}

			response := pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_product_file",
							apiPrefix,
							product.Slug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, response),
				),
			)
		})

		It("removes product file", func() {
			session := runMainWithArgs(
				"remove-product-file",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
				"--product-file-id", strconv.Itoa(productFile.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
		})
	})

	Describe("delete product-file", func() {
		var (
			productFile pivnet.ProductFile
		)

		BeforeEach(func() {
			productFile = pivnet.ProductFile{
				ID:   1234,
				Name: "some-product-file",
			}

			response := pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf(
							"%s/products/%s/product_files/%d",
							apiPrefix,
							product.Slug,
							productFile.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)
		})

		It("deletes product file", func() {
			session := runMainWithArgs(
				"delete-product-file",
				"--product-slug", product.Slug,
				"--product-file-id", strconv.Itoa(productFile.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
		})
	})
})
