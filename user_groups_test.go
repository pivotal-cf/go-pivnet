package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-golang/lager"
)

var _ = Describe("PivnetClient - user groups", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      lager.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = lager.NewLogger("user groups")
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List", func() {
		It("returns all user groups", func() {
			response := `{"user_groups": [{"id":2,"name":"group 1"},{"id": 3, "name": "group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			userGroups, err := client.UserGroups.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroups).To(HaveLen(2))
			Expect(userGroups[0].ID).To(Equal(2))
			Expect(userGroups[1].ID).To(Equal(3))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/user_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.UserGroups.List()
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("List for release", func() {
		var (
			releaseID int
		)

		BeforeEach(func() {
			releaseID = 1234
		})

		It("returns the user groups for the product slug", func() {
			response := `{"user_groups": [{"id":2,"name":"group 1"},{"id": 3, "name": "group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/banana/releases/%d/user_groups", apiPrefix, releaseID)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			userGroups, err := client.UserGroups.ListForRelease("banana", releaseID)
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroups).To(HaveLen(2))
			Expect(userGroups[0].ID).To(Equal(2))
			Expect(userGroups[1].ID).To(Equal(3))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/banana/releases/%d/user_groups", apiPrefix, releaseID)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.UserGroups.ListForRelease("banana", releaseID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Add", func() {
		var (
			productSlug = "banana-slug"
			releaseID   = 2345
			userGroupID = 3456

			expectedRequestBody = `{"user_group":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%s/releases/%d/add_user_group",
							apiPrefix,
							productSlug,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.UserGroups.AddToRelease(productSlug, releaseID, userGroupID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 204")))
			})
		})
	})

	Describe("Get User Group", func() {
		var (
			userGroupID int

			response           pivnet.UserGroup
			responseStatusCode int
		)

		BeforeEach(func() {
			userGroupID = 1234

			response = pivnet.UserGroup{
				ID:   userGroupID,
				Name: "something",
			}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/user_groups/%d",
							apiPrefix,
							userGroupID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the user group without error", func() {
			userGroup, err := client.UserGroups.Get(
				userGroupID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(userGroup.ID).To(Equal(userGroupID))
			Expect(userGroup.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.UserGroups.Get(
					userGroupID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Create", func() {
		var (
			name        string
			description string
			members     []string

			expectedRequestBody string

			returnedUserGroup pivnet.UserGroup
		)

		BeforeEach(func() {
			name = "some name"
			description = "some description"
			members = []string{"some member"}

			expectedRequestBody = fmt.Sprintf(
				`{"user_group":{"name":"%s","description":"%s","members":["some member"]}}`,
				name,
				description,
			)
		})

		JustBeforeEach(func() {
			returnedUserGroup = pivnet.UserGroup{
				ID:          1234,
				Name:        name,
				Description: description,
				Members:     members,
			}
		})

		Context("when members is nil", func() {
			BeforeEach(func() {
				members = nil

				expectedRequestBody = fmt.Sprintf(
					`{"user_group":{"name":"%s","description":"%s","members":[]}}`,
					name,
					description,
				)
			})

			It("successfully sends empty array in json body", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/user_groups",
							apiPrefix,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWithJSONEncoded(http.StatusCreated, returnedUserGroup),
					),
				)

				_, err := client.UserGroups.Create(name, description, members)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a 201 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/user_groups",
							apiPrefix,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWithJSONEncoded(http.StatusCreated, returnedUserGroup),
					),
				)

				userGroup, err := client.UserGroups.Create(name, description, members)
				Expect(err).NotTo(HaveOccurred())

				Expect(userGroup.ID).To(Equal(returnedUserGroup.ID))
				Expect(userGroup.Name).To(Equal(name))
				Expect(userGroup.Description).To(Equal(description))
			})
		})

		Context("when the server responds with a non-201 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", fmt.Sprintf(
							"%s/user_groups",
							apiPrefix,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.UserGroups.Create(name, description, members)

				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 201")))
			})
		})
	})
	Describe("Delete", func() {
		var (
			userGroup pivnet.UserGroup
		)

		BeforeEach(func() {
			userGroup = pivnet.UserGroup{
				ID: 1234,
			}
		})

		It("deletes the release", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroup.ID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)

			err := client.UserGroups.Delete(userGroup.ID)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-204 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/user_groups/%d", apiPrefix, userGroup.ID)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.UserGroups.Delete(userGroup.ID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 204")))
			})
		})
	})
})
