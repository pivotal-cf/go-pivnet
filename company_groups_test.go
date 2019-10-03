package pivnet_test

import (
	"fmt"
	"github.com/pivotal-cf/go-pivnet/v2/go-pivnetfakes"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet/v2"
	"github.com/pivotal-cf/go-pivnet/v2/logger"
	"github.com/pivotal-cf/go-pivnet/v2/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - company groups", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		apiAddress string
		userAgent  string

		newClientConfig        pivnet.ClientConfig
		fakeLogger             logger.Logger
		fakeAccessTokenService *gopivnetfakes.FakeAccessTokenService
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &loggerfakes.FakeLogger{}
		fakeAccessTokenService = &gopivnetfakes.FakeAccessTokenService{}
		newClientConfig = pivnet.ClientConfig{
			Host:      apiAddress,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(fakeAccessTokenService, newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("List", func() {
		It("returns all company groups", func() {
			response := `{"company_groups": [{"id":2,"name":"company group 1"},{"id": 3, "name": "company group 2"}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/company_groups", apiPrefix)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			companyGroups, err := client.CompanyGroups.List()
			Expect(err).NotTo(HaveOccurred())

			Expect(companyGroups).To(HaveLen(2))
			Expect(companyGroups[0].ID).To(Equal(2))
			Expect(companyGroups[1].ID).To(Equal(3))
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
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/company_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, body),
					),
				)

				_, err := client.CompanyGroups.List()
				Expect(err.Error()).To(ContainSubstring("foo message"))
			})
		})

		Context("when the json unmarshalling fails with error", func() {
			It("forwards the error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", fmt.Sprintf("%s/company_groups", apiPrefix)),
						ghttp.RespondWith(http.StatusTeapot, "%%%"),
					),
				)

				_, err := client.CompanyGroups.List()
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("invalid character"))
			})
		})
	})
})
