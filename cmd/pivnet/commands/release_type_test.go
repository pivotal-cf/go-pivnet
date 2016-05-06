package commands_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("release types commands", func() {
	var (
		server *ghttp.Server

		outBuffer bytes.Buffer

		releaseTypes []string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.StdOutWriter = &outBuffer

		releaseTypes = []string{
			"release type 1",
			"release type 2",
		}
	})

	AfterEach(func() {
		server.Close()
	})

	It("lists all release types", func() {
		releaseTypesResponse := pivnet.ReleaseTypesResponse{
			ReleaseTypes: releaseTypes,
		}

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
				ghttp.RespondWithJSONEncoded(http.StatusOK, releaseTypesResponse),
			),
		)

		releaseTypesCommand := commands.ReleaseTypesCommand{}

		err := releaseTypesCommand.Execute(nil)
		Expect(err).NotTo(HaveOccurred())

		var returnedReleaseTypes []string

		err = json.Unmarshal(outBuffer.Bytes(), &returnedReleaseTypes)
		Expect(err).NotTo(HaveOccurred())

		Expect(returnedReleaseTypes).To(Equal(releaseTypes))
	})
})
