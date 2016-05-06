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

var _ = Describe("file group commands", func() {
	var (
		server *ghttp.Server

		field     reflect.StructField
		outBuffer bytes.Buffer

		productSlug string

		release    pivnet.Release
		releases   []pivnet.Release
		fileGroups []pivnet.FileGroup
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

		fileGroups = []pivnet.FileGroup{
			{
				ID:   1234,
				Name: "Some file group",
			},
			{
				ID:   2345,
				Name: "Another file group",
			},
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("FileGroupsCommand", func() {
		It("lists all file groups for the provided product slug", func() {
			fileGroupsResponse := pivnet.FileGroupsResponse{
				FileGroups: fileGroups,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/file_groups",
							apiPrefix,
							productSlug,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, fileGroupsResponse),
				),
			)

			fileGroupsCommand := commands.FileGroupsCommand{
				ProductSlug: productSlug,
			}

			err := fileGroupsCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returned []pivnet.FileGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returned)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned).To(Equal(fileGroups))
		})

		Context("when release version is provided", func() {
			It("lists all file groups for the provided product slug and release version", func() {
				releasesResponse := pivnet.ReleasesResponse{
					Releases: releases,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
						ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
					),
				)

				fileGroupsResponse := pivnet.FileGroupsResponse{
					FileGroups: fileGroups,
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							fmt.Sprintf(
								"%s/products/%s/releases/%d/file_groups",
								apiPrefix,
								productSlug,
								releases[0].ID,
							),
						),
						ghttp.RespondWithJSONEncoded(http.StatusOK, fileGroupsResponse),
					),
				)

				fileGroupsCommand := commands.FileGroupsCommand{
					ProductSlug:    productSlug,
					ReleaseVersion: releases[0].Version,
				}

				err := fileGroupsCommand.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returned []pivnet.FileGroup

				err = json.Unmarshal(outBuffer.Bytes(), &returned)
				Expect(err).NotTo(HaveOccurred())

				Expect(returned).To(Equal(fileGroups))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupsCommand{}, "ProductSlug")
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
				field = fieldFor(commands.FileGroupsCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("FileGroupCommand", func() {
		It("shows the file group for the provided product slug and file group id", func() {
			fileGroupResponse := fileGroups[0]

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/file_groups/%d",
							apiPrefix,
							productSlug,
							fileGroups[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, fileGroupResponse),
				),
			)

			fileGroupCommand := commands.FileGroupCommand{
				ProductSlug: productSlug,
				FileGroupID: fileGroups[0].ID,
			}

			err := fileGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returned pivnet.FileGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returned)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned).To(Equal(fileGroups[0]))
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("FileGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupCommand{}, "FileGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("file-group-id"))
			})
		})
	})

	Describe("DeleteFileGroupCommand", func() {
		It("deletes the file group for the provided product slug and file group id", func() {
			fileGroupResponse := fileGroups[0]

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf(
							"%s/products/%s/file_groups/%d",
							apiPrefix,
							productSlug,
							fileGroups[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, fileGroupResponse),
				),
			)

			deleteFileGroupCommand := commands.DeleteFileGroupCommand{}
			deleteFileGroupCommand.ProductSlug = productSlug
			deleteFileGroupCommand.FileGroupID = fileGroups[0].ID

			err := deleteFileGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteFileGroupCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("FileGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupCommand{}, "FileGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("file-group-id"))
			})
		})
	})
})
