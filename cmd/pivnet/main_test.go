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

		product  pivnet.Product
		products []pivnet.Product

		eulas []pivnet.EULA

		releases []pivnet.Release
		release  pivnet.Release
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		host = server.URL()

		product = pivnet.Product{
			ID:   1234,
			Slug: "some-product-slug",
			Name: "some-product-name",
		}

		products = []pivnet.Product{
			product,
			{
				ID:   2345,
				Slug: "another-product-slug",
				Name: "another-product-name",
			},
		}

		eulas = []pivnet.EULA{
			{
				ID:   1234,
				Name: "some eula",
				Slug: "some-eula",
			},
			{
				ID:   2345,
				Name: "another eula",
				Slug: "another-eula",
			},
		}

		release = pivnet.Release{
			ID:          1234,
			Version:     "version 0.2.3",
			Description: "Some release with some description.",
		}

		releases = []pivnet.Release{
			release,
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
			session := runMainWithArgs("--format=json", "product", "-s", product.Slug)

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
			session := runMainWithArgs("--format=yaml", "product", "-s", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))

			var receivedProduct pivnet.Product
			err := yaml.Unmarshal(session.Out.Contents(), &receivedProduct)
			Expect(err).NotTo(HaveOccurred())

			Expect(receivedProduct.Slug).To(Equal(product.Slug))
		})
	})

	Describe("product", func() {
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

		It("displays product for the provided slug", func() {
			session := runMainWithArgs("product", "-s", product.Slug)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(product.Slug))
		})
	})

	Describe("products", func() {
		BeforeEach(func() {
			response := pivnet.ProductsResponse{
				Products: products,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products", apiPrefix),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)
		})

		It("displays products", func() {
			session := runMainWithArgs("products")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(product.Slug))
		})
	})

	Describe("EULAs", func() {
		It("displays all EULAs", func() {
			eulasResponse := pivnet.EULAsResponse{
				EULAs: eulas,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, eulasResponse),
				),
			)

			session := runMainWithArgs("eulas")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(eulas[0].Name))
			Expect(session).Should(gbytes.Say(eulas[1].Name))
		})

		It("accepts EULAs", func() {
			eulaAcceptanceResponse := pivnet.EULAAcceptanceResponse{
				AcceptedAt: "now",
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"POST",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/eula_acceptance",
							apiPrefix,
							product.Slug,
							release.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, eulaAcceptanceResponse),
				),
			)

			session := runMainWithArgs(
				"accept-eula",
				"--product-slug", product.Slug,
				"--release-id", strconv.Itoa(release.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
		})
	})

	Describe("Release Types", func() {
		var (
			releaseTypes []string
		)

		BeforeEach(func() {
			releaseTypes = []string{"some release type", "another release type"}
		})

		It("displays all release types", func() {
			releaseTypesResponse := pivnet.ReleaseTypesResponse{
				ReleaseTypes: releaseTypes,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseTypesResponse),
				),
			)

			session := runMainWithArgs("release-types")

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(releaseTypes[0]))
			Expect(session).Should(gbytes.Say(releaseTypes[1]))
		})
	})

	Describe("Releases", func() {
		It("displays all releases for the provided product slug", func() {
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

			session := runMainWithArgs(
				"releases",
				"--product-slug", product.Slug,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(releases[0].Version))
			Expect(session).Should(gbytes.Say(releases[1].Version))
		})

		It("displays release for the provided product slug and release version", func() {
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

			session := runMainWithArgs(
				"release",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(releases[0].Version))
		})

		It("displays release for the provided product slug and release version", func() {
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

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf(
							"%s/products/%s/releases/%d",
							apiPrefix,
							product.Slug,
							release.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, nil),
				),
			)

			session := runMainWithArgs(
				"delete-release",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
		})
	})

	Describe("User groups", func() {
		It("displays user groups for the provided product slug and release version", func() {
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

			userGroups := []pivnet.UserGroup{
				{
					ID:   1234,
					Name: "Some user group",
				},
				{
					ID:   2345,
					Name: "Another user group",
				},
			}

			userGroupsResponse := pivnet.UserGroups{
				UserGroups: userGroups,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/user_groups",
							apiPrefix,
							product.Slug,
							release.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, userGroupsResponse),
				),
			)

			session := runMainWithArgs(
				"user-groups",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(userGroups[0].Name))
		})
	})

	Describe("Release dependencies", func() {
		It("displays release dependencies for the provided product slug and release version", func() {
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

			releaseDependencies := []pivnet.ReleaseDependency{
				{
					Release: pivnet.DependentRelease{
						ID:      1234,
						Version: "Some version",
					},
				},
				{
					Release: pivnet.DependentRelease{
						ID:      2345,
						Version: "Another version",
					},
				},
			}

			releaseDependenciesResponse := pivnet.ReleaseDependenciesResponse{
				ReleaseDependencies: releaseDependencies,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/dependencies",
							apiPrefix,
							product.Slug,
							release.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseDependenciesResponse),
				),
			)

			session := runMainWithArgs(
				"release-dependencies",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(releaseDependencies[0].Release.Version))
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

			response := pivnet.ProductFileResponse{
				ProductFile: productFile,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/products/%s/releases/%d/product_files/%d",
							apiPrefix, product.Slug, release.ID, productFile.ID),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)
		})

		It("displays product file", func() {
			session := runMainWithArgs(
				"product-file",
				"--product-slug", product.Slug,
				"--release-id", strconv.Itoa(release.ID),
				"--product-file-id", strconv.Itoa(productFile.ID),
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(productFile.Name))
		})
	})

	Describe("Release upgrade paths", func() {
		It("displays release upgrade paths for the provided product slug and release version", func() {
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

			releaseUpgradePaths := []pivnet.ReleaseUpgradePath{
				{
					Release: pivnet.UpgradePathRelease{
						ID:      1234,
						Version: "Some version",
					},
				},
				{
					Release: pivnet.UpgradePathRelease{
						ID:      2345,
						Version: "Another version",
					},
				},
			}

			releaseUpgradePathsResponse := pivnet.ReleaseUpgradePathsResponse{
				ReleaseUpgradePaths: releaseUpgradePaths,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/upgrade_paths",
							apiPrefix,
							product.Slug,
							release.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseUpgradePathsResponse),
				),
			)

			session := runMainWithArgs(
				"release-upgrade-paths",
				"--product-slug", product.Slug,
				"--release-version", releases[0].Version,
			)

			Eventually(session, executableTimeout).Should(gexec.Exit(0))
			Expect(session).Should(gbytes.Say(releaseUpgradePaths[0].Release.Version))
		})
	})
})
