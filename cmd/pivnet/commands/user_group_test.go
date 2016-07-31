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
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("user group commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		field     reflect.StructField
		outBuffer bytes.Buffer

		userGroups []pivnet.UserGroup
		userGroup  pivnet.UserGroup

		releases []pivnet.Release

		productSlug string

		responseStatusCode int
		response           interface{}

		releasesResponseStatusCode int
		releasesResponse           pivnet.ReleasesResponse
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}
		commands.ErrorHandler = fakeErrorHandler

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

		releasesResponseStatusCode = http.StatusOK

		releasesResponse = pivnet.ReleasesResponse{
			Releases: releases,
		}

		responseStatusCode = http.StatusOK
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("UserGroupsCommand", func() {
		var (
			command commands.UserGroupsCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusOK

			command = commands.UserGroupsCommand{}

			response = pivnet.UserGroupsResponse{
				userGroups,
			}
		})

		Describe("All user groups", func() {
			JustBeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
						ghttp.RespondWithJSONEncoded(responseStatusCode, response),
					),
				)
			})

			It("lists all user groups", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returnedUserGroups []pivnet.UserGroup

				err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroups)
				Expect(err).NotTo(HaveOccurred())

				Expect(returnedUserGroups).To(Equal(userGroups))
			})

			Context("when there is an error", func() {
				BeforeEach(func() {
					responseStatusCode = http.StatusTeapot
				})

				It("invokes the error handler", func() {
					err := command.Execute(nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				})
			})
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

				command = commands.UserGroupsCommand{
					ProductSlug:    productSlug,
					ReleaseVersion: release.Version,
				}
			})

			JustBeforeEach(func() {
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
						ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
					),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf(
							"%s/products/%s/releases/%d/user_groups",
							apiPrefix,
							productSlug,
							release.ID,
						)),
						ghttp.RespondWithJSONEncoded(responseStatusCode, response),
					),
				)
			})

			It("displays user groups for the provided product slug and release version", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				var returnedUserGroups []pivnet.UserGroup

				err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroups)
				Expect(err).NotTo(HaveOccurred())

				Expect(returnedUserGroups).To(Equal(userGroups))
			})

			Context("when there is an error", func() {
				BeforeEach(func() {
					responseStatusCode = http.StatusTeapot
				})

				It("invokes the error handler", func() {
					err := command.Execute(nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				})
			})

			Context("when there is an error getting all releases", func() {
				BeforeEach(func() {
					releasesResponseStatusCode = http.StatusTeapot
				})

				It("invokes the error handler", func() {
					err := command.Execute(nil)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				})
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

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
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

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("UserGroupCommand", func() {
		var (
			command commands.UserGroupCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusOK

			command = commands.UserGroupCommand{
				UserGroupID: userGroup.ID,
			}

			response = userGroup
		})

		JustBeforeEach(func() {
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
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("shows the user group for user group id", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returned pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returned)
			Expect(err).NotTo(HaveOccurred())

			Expect(returned).To(Equal(userGroup))
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})
	})

	Describe("CreateUserGroupCommand", func() {
		var (
			command commands.CreateUserGroupCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusCreated

			command = commands.CreateUserGroupCommand{
				Name:        "some name",
				Description: "some description",
			}

			response = userGroups[0]
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("%s/user_groups", apiPrefix)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("creates user group", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedUserGroup pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedUserGroup).To(Equal(userGroups[0]))
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
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

			userGroupsResponseStatusCode int

			updatedUserGroup           pivnet.UserGroup
			updateUserGroupRequestBody string

			command commands.UpdateUserGroupCommand
		)

		BeforeEach(func() {
			nameVal := "updated name"
			descriptionVal := "updated description"

			name = &nameVal
			description = &descriptionVal

			updateUserGroupRequestBody = fmt.Sprintf(
				`{"user_group":{"name":"%s","description":"%s"}}`,
				*name,
				*description,
			)

			updatedUserGroup = userGroups[0]
			updatedUserGroup.Name = *name
			updatedUserGroup.Description = *description

			command = commands.UpdateUserGroupCommand{
				UserGroupID: userGroups[0].ID,
				Name:        name,
				Description: description,
			}

			userGroupsResponseStatusCode = http.StatusOK

			responseStatusCode = http.StatusOK

			response = pivnet.UpdateUserGroupResponse{
				updatedUserGroup,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroups[0].ID)),
					ghttp.RespondWithJSONEncoded(userGroupsResponseStatusCode, userGroups[0]),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroups[0].ID)),
					ghttp.VerifyJSON(updateUserGroupRequestBody),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("updates name and description for user group", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedUserGroup pivnet.UserGroup

			err = json.Unmarshal(outBuffer.Bytes(), &returnedUserGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedUserGroup).To(Equal(updatedUserGroup))
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Context("when there is an error getting all user groups", func() {
			BeforeEach(func() {
				userGroupsResponseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Context("when name is not provided", func() {
			BeforeEach(func() {
				updateUserGroupRequestBody = fmt.Sprintf(
					`{"user_group":{"name":"%s","description":"%s"}}`,
					userGroups[0].Name,
					*description,
				)

				command.Name = nil
			})

			It("uses previous name in request body", func() {
				err := command.Execute(nil)
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

				command.Description = nil
			})

			It("uses previous description in request body", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())
			})
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
		var (
			command commands.AddUserGroupCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			command = commands.AddUserGroupCommand{
				ProductSlug:    productSlug,
				UserGroupID:    userGroup.ID,
				ReleaseVersion: releases[0].Version,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
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
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("adds the user group for the provided product slug and user group id to the specified release", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Context("when there is an error getting all releases", func() {
			BeforeEach(func() {
				releasesResponseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
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

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("RemoveUserGroupCommand", func() {
		var (
			command commands.RemoveUserGroupCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			command = commands.RemoveUserGroupCommand{
				ProductSlug:    productSlug,
				UserGroupID:    userGroup.ID,
				ReleaseVersion: releases[0].Version,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/releases", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(releasesResponseStatusCode, releasesResponse),
				),
			)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/remove_user_group",
							apiPrefix,
							productSlug,
							releases[0].ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("removes the user group for the provided product slug and user group id from the specified release", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Context("when there is an error getting all releases", func() {
			BeforeEach(func() {
				releasesResponseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveUserGroupCommand{}, "ProductSlug")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("p"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-slug"))
			})
		})

		Describe("UserGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveUserGroupCommand{}, "UserGroupID")
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
				field = fieldFor(commands.RemoveUserGroupCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("v"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("DeleteUserGroupCommand", func() {
		var (
			command commands.DeleteUserGroupCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			command = commands.DeleteUserGroupCommand{
				UserGroupID: userGroup.ID,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroup.ID)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("deletes user group", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
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

	Describe("AddUserGroupMemberCommand", func() {
		var (
			command commands.AddUserGroupMemberCommand

			memberEmailAddress string
			admin              bool
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			memberEmailAddress = "some email address"
			admin = true

			command = commands.AddUserGroupMemberCommand{
				UserGroupID:        userGroup.ID,
				MemberEmailAddress: memberEmailAddress,
				Admin:              admin,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/user_groups/%d/add_member",
							apiPrefix,
							userGroup.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("adds member", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("UserGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupMemberCommand{}, "UserGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("user-group-id"))
			})
		})

		Describe("MemberEmailAddress flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupMemberCommand{}, "MemberEmailAddress")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("member-email"))
			})
		})

		Describe("Admin flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddUserGroupMemberCommand{}, "Admin")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("admin"))
			})
		})
	})

	Describe("RemoveUserGroupMemberCommand", func() {
		var (
			command commands.RemoveUserGroupMemberCommand

			memberEmailAddress string
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			memberEmailAddress = "some email address"
			command = commands.RemoveUserGroupMemberCommand{
				UserGroupID:        userGroup.ID,
				MemberEmailAddress: memberEmailAddress,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/user_groups/%d/remove_member",
							apiPrefix,
							userGroup.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("removes member", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("UserGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveUserGroupMemberCommand{}, "UserGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("user-group-id"))
			})
		})

		Describe("MemberEmailAddress flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveUserGroupMemberCommand{}, "MemberEmailAddress")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("member-email"))
			})
		})
	})
	Describe("RemoveUserGroupMemberCommand", func() {
		var (
			command commands.RemoveUserGroupMemberCommand

			memberEmailAddress string
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusNoContent

			memberEmailAddress = "some email address"
			command = commands.RemoveUserGroupMemberCommand{
				UserGroupID:        userGroup.ID,
				MemberEmailAddress: memberEmailAddress,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"PATCH",
						fmt.Sprintf(
							"%s/user_groups/%d/remove_member",
							apiPrefix,
							userGroup.ID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, nil),
				),
			)
		})

		It("deletes user group", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("invokes the error handler", func() {
				err := command.Execute(nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
			})
		})

		Describe("UserGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveUserGroupMemberCommand{}, "UserGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("user-group-id"))
			})
		})

		Describe("MemberEmailAddress flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveUserGroupMemberCommand{}, "MemberEmailAddress")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("member-email"))
			})
		})
	})
})
