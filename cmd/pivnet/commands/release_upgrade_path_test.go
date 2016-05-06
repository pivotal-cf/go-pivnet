package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("release upgrade path commands", func() {
	var (
		server *ghttp.Server

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string

		release             pivnet.Release
		releases            []pivnet.Release
		releaseUpgradePaths []pivnet.ReleaseUpgradePath
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer

		productSlug = "some-product-slug"

		release = pivnet.Release{
			ID:      1234,
			Version: "some-release-version",
		}

		releases = []pivnet.Release{
			release,
			{
				ID:      2345,
				Version: "another-release-version",
			},
		}

		releaseUpgradePaths = []pivnet.ReleaseUpgradePath{
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
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleasesUpgradePathsCommand", func() {
		It("lists all release dependencies for the provided product slug and release version", func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

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
							productSlug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releaseUpgradePathsResponse),
				),
			)

			releaseUpgradePathsCommand := commands.ReleaseUpgradePathsCommand{}
			releaseUpgradePathsCommand.ProductSlug = productSlug
			releaseUpgradePathsCommand.ReleaseVersion = releases[0].Version

			err := releaseUpgradePathsCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returned []pivnet.ReleaseUpgradePath

			err = json.Unmarshal(outBuffer.Bytes(), &returned)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned).To(Equal(releaseUpgradePaths))
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseUpgradePathsCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ReleaseUpgradePathsCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})
})
