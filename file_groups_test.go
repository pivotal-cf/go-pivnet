package pivnet_test

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
	"github.com/pivotal-cf-experimental/go-pivnet/logger/loggerfakes"
)

var _ = Describe("PivnetClient - FileGroup", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
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
		It("returns all FileGroups", func() {
			response := pivnet.FileGroupsResponse{
				[]pivnet.FileGroup{
					{
						ID:   1234,
						Name: "Some file group",
					},
					{
						ID:   2345,
						Name: "Some other file group",
					},
				},
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/file_groups", apiPrefix, productSlug)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, response),
				),
			)

			fileGroups, err := client.FileGroups.List(productSlug)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroups).To(HaveLen(2))

			Expect(fileGroups[0].ID).To(Equal(fileGroups[0].ID))
			Expect(fileGroups[0].Name).To(Equal(fileGroups[0].Name))
			Expect(fileGroups[1].ID).To(Equal(fileGroups[1].ID))
			Expect(fileGroups[1].Name).To(Equal(fileGroups[1].Name))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/products/%s/file_groups", apiPrefix, productSlug)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.FileGroups.List(productSlug)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("List for release", func() {
		var (
			productSlug string
			releaseID   int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			releaseID = 12

			response = pivnet.FileGroupsResponse{[]pivnet.FileGroup{
				{
					ID:   1234,
					Name: "something",
				},
				{
					ID:   2345,
					Name: "something-else",
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						fmt.Sprintf(
							"%s/products/%s/releases/%d/file_groups",
							apiPrefix,
							productSlug,
							releaseID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			fileGroups, err := client.FileGroups.ListForRelease(
				productSlug,
				releaseID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroups).To(HaveLen(2))
			Expect(fileGroups[0].ID).To(Equal(1234))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.FileGroups.ListForRelease(
					productSlug,
					releaseID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Get File group", func() {
		var (
			productSlug string
			fileGroupID int

			response           interface{}
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			fileGroupID = 1234

			response = pivnet.FileGroup{
				ID:   fileGroupID,
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
							"%s/products/%s/file_groups/%d",
							apiPrefix,
							productSlug,
							fileGroupID,
						),
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the file group without error", func() {
			fileGroup, err := client.FileGroups.Get(
				productSlug,
				fileGroupID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroup.ID).To(Equal(fileGroupID))
			Expect(fileGroup.Name).To(Equal("something"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
				response = pivnetErr{Message: "foo message"}
			})

			It("returns an error", func() {
				_, err := client.FileGroups.Get(
					productSlug,
					fileGroupID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})

	Describe("Delete File Group", func() {
		var (
			id = 1234
		)

		It("deletes the file group", func() {
			response := []byte(`{"id":1234}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/file_groups/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			fileGroup, err := client.FileGroups.Delete(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileGroup.ID).To(Equal(id))
		})

		Context("when the server responds with a non-2XX status code", func() {
			var (
				body []byte
			)

			BeforeEach(func() {
				body = []byte(`{"message":"foo message"}`)
			})

			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/file_groups/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.FileGroups.Delete(productSlug, id)
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})
	})
})
