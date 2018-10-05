package pivnet_test

import (
	"fmt"
	"net/http"

	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PivnetClient - federation token", func() {
	type requestBody struct {
		ProductID string `json:"product_id"`
	}

	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		response           interface{}
		responseStatusCode int
		expectedRequestBody requestBody
		newClientConfig pivnet.ClientConfig
		fakeLogger      logger.Logger

		productSlug string
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

	JustBeforeEach(func() {
		expectedRequestBody = requestBody{
			ProductID: productSlug,
		}

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(
					"POST",
					fmt.Sprintf("%s/federation_token", apiPrefix),
				),
				ghttp.VerifyJSONRepresenting(&expectedRequestBody),
				ghttp.RespondWithJSONEncoded(responseStatusCode, response),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Generate a federated token", func() {
		BeforeEach(func() {
			productSlug = "banana"

			response = pivnet.FederationToken{
				AccessKeyID: "some-AccessKeyID",
				SecretAccessKey: "some-SecretAccessKey",
				SessionToken: "some-SessionToken",
			}

			responseStatusCode = http.StatusOK
		})

		It("returns the federated token without error", func() {
			federationToken, err := client.FederationToken.GenerateFederationToken(
				productSlug,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(federationToken).ToNot(Equal(nil))
			Expect(federationToken.AccessKeyID).To(Equal("some-AccessKeyID"))
			Expect(federationToken.SecretAccessKey).To(Equal("some-SecretAccessKey"))
			Expect(federationToken.SessionToken).To(Equal("some-SessionToken"))
		})
	})

	Describe("Err when trying to generate token for restricted product", func() {
		BeforeEach(func() {
			productSlug = "something-i-dont-manage"

			response = pivnetErr{Message: "only available for product admins and partner product admins"}

			responseStatusCode = http.StatusForbidden
		})

		It("returns a 403 error", func() {
			federationToken, err := client.FederationToken.GenerateFederationToken(
				productSlug,
			)

			Expect(federationToken).To(Equal(pivnet.FederationToken{}))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("403 - only available for product admins and partner product admins"))
		})
	})
})