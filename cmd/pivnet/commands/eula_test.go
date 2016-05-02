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

var _ = Describe("eula commands", func() {
	var (
		server *ghttp.Server
		host   string

		eulas []pivnet.EULA

		outBuffer bytes.Buffer
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		host = server.URL()
		commands.Pivnet.Host = host
		outBuffer = bytes.Buffer{}
		commands.OutWriter = &outBuffer

		eulas = []pivnet.EULA{
			{
				ID:   1234,
				Name: "some eula",
				Slug: "some-eula",
			},
			{
				ID:   2345,
				Name: "another eula",
				Slug: "another-eula",
			},
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("listing EULAs", func() {
		It("lists all EULAs", func() {
			eulasResponse := pivnet.EULAsResponse{
				EULAs: eulas,
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas", apiPrefix)),
					ghttp.RespondWithJSONEncoded(http.StatusOK, eulasResponse),
				),
			)

			eulasCommand := commands.EULAsCommand{}
			err := eulasCommand.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedEULAs []pivnet.EULA

			err = json.Unmarshal(outBuffer.Bytes(), &returnedEULAs)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedEULAs).To(Equal(eulas))
		})
	})

	It("shows specific EULA", func() {
		eulaResponse := eulas[0]

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf("%s/eulas/%s", apiPrefix, eulas[0].Slug)),
				ghttp.RespondWithJSONEncoded(http.StatusOK, eulaResponse),
			),
		)

		eulaCommand := commands.EULACommand{}
		eulaCommand.EULASlug = eulas[0].Slug
		err := eulaCommand.Execute(nil)
		Expect(err).NotTo(HaveOccurred())

		var returnedEULA pivnet.EULA

		err = json.Unmarshal(outBuffer.Bytes(), &returnedEULA)
		Expect(err).NotTo(HaveOccurred())

		Expect(returnedEULA).To(Equal(eulas[0]))
	})
})
