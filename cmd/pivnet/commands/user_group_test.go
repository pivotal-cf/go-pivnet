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

var _ = Describe("user group commands", func() {
	var (
		server *ghttp.Server

		field     reflect.StructField
		outBuffer bytes.Buffer

		userGroups []pivnet.UserGroup
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutWriter = &outBuffer

		userGroups = []pivnet.UserGroup{
			{
				ID:   1234,
				Name: "Some user group",
			},
			{
				ID:   2345,
				Name: "Another user group",
			},
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("UserGroupsCommand", func() {
		It("lists all user groups", func() {
			userGroupsResponse := pivnet.UserGroupsResponse{userGroups}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, userGroupsResponse),
				),
			)

			userGroupsCommand := commands.UserGroupsCommand{}

			err := userGroupsCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedUserGroups []pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroups)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedUserGroups).To(Equal(userGroups))
		})

		Context("when product slug and release version are provided", func() {
			var (
				productSlug string
				release     pivnet.Release
			)

			BeforeEach(func() {
				productSlug = "some-product-slug"
				release = pivnet.Release{
					ID:      1234,
					Version: "some-release-version",
				}
			})

			It("displays user groups for the provided product slug and release version", func() {
				releasesResponse := pivnet.ReleasesResponse{
					Releases: []pivnet.Release{
						release,
					},
				}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases",
							apiPrefix,
							productSlug,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
					),
				)

				userGroupsResponse := pivnet.UserGroupsResponse{userGroups}

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/user_groups",
							apiPrefix,
							productSlug,
							release.ID,
						)),
						ghttp.RespondWithJSONEncoded(http.StatusOK, userGroupsResponse),
					),
				)

				userGroupsCommand := commands.UserGroupsCommand{}
				userGroupsCommand.ProductSlug = productSlug
				userGroupsCommand.ReleaseVersion = release.Version

				err := userGroupsCommand.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returnedUserGroups []pivnet.UserGroup

				err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroups)
				Expect(err).NotTo(HaveOccurred())

				Expect(returnedUserGroups).To(Equal(userGroups))
			})
		})

		Context("when only product slug is provided", func() {
			It("returns an error", func() {
				userGroupsCommand := commands.UserGroupsCommand{}
				userGroupsCommand.ProductSlug = "some-slug"

				err := userGroupsCommand.Execute(nil)
				Expect(err).To(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(0))

			})
		})

		Context("when only release version is provided", func() {
			It("returns an error", func() {
				userGroupsCommand := commands.UserGroupsCommand{}
				userGroupsCommand.ReleaseVersion = "some-version"

				err := userGroupsCommand.Execute(nil)
				Expect(err).To(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(0))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.UserGroupsCommand{}, "ProductSlug")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.UserGroupsCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("CreateUserGroupCommand", func() {
		It("creates user group", func() {
			createUserGroupResponse := userGroups[0]

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("%s/user_groups", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusCreated, createUserGroupResponse),
				),
			)

			createUserGroupCommand := commands.CreateUserGroupCommand{}
			createUserGroupCommand.Name = "some name"
			createUserGroupCommand.Description = "some description"

			err := createUserGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedUserGroup pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedUserGroup).To(Equal(userGroups[0]))
		})

		Describe("Name flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.CreateUserGroupCommand{}, "Name")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("name"))
			})
		})

		Describe("Description flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.CreateUserGroupCommand{}, "Description")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("description"))
			})
		})

		Describe("Members flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.CreateUserGroupCommand{}, "Members")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("member"))
			})
		})
	})
})
