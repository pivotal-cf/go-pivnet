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
		userGroup  pivnet.UserGroup

		releases []pivnet.Release

		productSlug string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer

		userGroups = []pivnet.UserGroup{
			{
				ID:          1234,
				Name:        "Some-user-group",
				Description: "Some user group",
			},
			{
				ID:          2345,
				Name:        "Another-user-group",
				Description: "Another user group",
			},
		}
		userGroup = userGroups[1]

		releases = []pivnet.Release{
			{
				ID:      1234,
				Version: "some-release-version",
			},
			{
				ID:      2345,
				Version: "another-release-version",
			},
		}

		productSlug = "some-fake-product-slug"
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

				userGroupsCommand := commands.UserGroupsCommand{
					ProductSlug:    productSlug,
					ReleaseVersion: release.Version,
				}

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
				userGroupsCommand := commands.UserGroupsCommand{
					ProductSlug: "some-slug",
				}

				err := userGroupsCommand.Execute(nil)
				Expect(err).To(HaveOccurred())

				Expect(server.ReceivedRequests()).To(HaveLen(0))

			})
		})

		Context("when only release version is provided", func() {
			It("returns an error", func() {
				userGroupsCommand := commands.UserGroupsCommand{
					ReleaseVersion: "some-version",
				}

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

	Describe("UserGroupCommand", func() {
		It("shows the user group for user group id", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/user_groups/%d",
							apiPrefix,
							userGroup.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusOK, userGroup),
				),
			)

			userGroupCommand := commands.UserGroupCommand{
				UserGroupID: userGroup.ID,
			}

			err := userGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returned pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returned)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned).To(Equal(userGroup))
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

			createUserGroupCommand := commands.CreateUserGroupCommand{
				Name:        "some name",
				Description: "some description",
			}

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

	Describe("UpdateUserGroupCommand", func() {
		var (
			name        *string
			description *string

			updatedUserGroup           pivnet.UserGroup
			updateUserGroupRequestBody string

			updateUserGroupCommand commands.UpdateUserGroupCommand
		)

		BeforeEach(func() {
			nameVal := "updated name"
			descriptionVal := "updated description"

			name = &nameVal
			description = &descriptionVal

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroups[0].ID)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, userGroups[0]),
				),
			)

			updateUserGroupRequestBody = fmt.Sprintf(
				`{"user_group":{"name":"%s","description":"%s"}}`,
				*name,
				*description,
			)

			updatedUserGroup = userGroups[0]
			updatedUserGroup.Name = *name
			updatedUserGroup.Description = *description

			updateUserGroupCommand = commands.UpdateUserGroupCommand{
				UserGroupID: userGroups[0].ID,
				Name:        name,
				Description: description,
			}
		})

		JustBeforeEach(func() {
			updateUserGroupResponse := pivnet.UpdateUserGroupResponse{updatedUserGroup}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroups[0].ID)),
					ghttp.VerifyJSON(updateUserGroupRequestBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, updateUserGroupResponse),
				),
			)
		})

		It("updates name and description for user group", func() {
			err := updateUserGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedUserGroup pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedUserGroup).To(Equal(updatedUserGroup))
		})

		Context("when name is not provided", func() {
			BeforeEach(func() {
				updateUserGroupRequestBody = fmt.Sprintf(
					`{"user_group":{"name":"%s","description":"%s"}}`,
					userGroups[0].Name,
					*description,
				)

				updateUserGroupCommand.Name = nil
			})

			It("uses previous name in request body", func() {
				err := updateUserGroupCommand.Execute(nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when description is not provided", func() {
			BeforeEach(func() {
				updateUserGroupRequestBody = fmt.Sprintf(
					`{"user_group":{"name":"%s","description":"%s"}}`,
					*name,
					userGroups[0].Description,
				)

				updateUserGroupCommand.Description = nil
			})

			It("uses previous description in request body", func() {
				err := updateUserGroupCommand.Execute(nil)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when description is empty", func() {

		})

		Describe("Name flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.UpdateUserGroupCommand{}, "Name")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("name"))
			})
		})

		Describe("Description flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.UpdateUserGroupCommand{}, "Description")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("description"))
			})
		})
	})

	Describe("AddUserGroupCommand", func() {
		It("adds the user group for the provided product slug and user group id to the specified release", func() {
			releasesResponse := pivnet.ReleasesResponse{
				Releases: releases,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/add_user_group",
							apiPrefix,
							productSlug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, nil),
				),
			)

			userGroupCommand := commands.AddUserGroupCommand{
				ProductSlug:    productSlug,
				UserGroupID:    userGroup.ID,
				ReleaseVersion: releases[0].Version,
			}

			err := userGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("UserGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupCommand{}, "UserGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("user-group-id"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("DeleteUserGroupCommand", func() {
		It("deletes user group", func() {
			userGroupID := 1234

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroupID)),
					ghttp.RespondWithJSONEncoded(http.StatusNoContent, nil),
				),
			)

			deleteUserGroupCommand := commands.DeleteUserGroupCommand{
				UserGroupID: userGroupID,
			}

			err := deleteUserGroupCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("UserGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteUserGroupCommand{}, "UserGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("user-group-id"))
			})
		})
	})
})
