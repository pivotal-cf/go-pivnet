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
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errors/errorsfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"

	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("release types commands", func() {
	var (
		server *ghttp.Server

		fakeErrorHandler *errorsfakes.FakeErrorHandler

		outBuffer bytes.Buffer

		releaseTypes []string

		responseStatusCode int
		response           interface{}
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		commands.Pivnet.Host = server.URL()

		outBuffer = bytes.Buffer{}
		commands.OutputWriter = &outBuffer
		commands.Printer = printer.NewPrinter(commands.OutputWriter)

		fakeErrorHandler = &errorsfakes.FakeErrorHandler{}
		commands.ErrorHandler = fakeErrorHandler

		releaseTypes = []string{
			"release type 1",
			"release type 2",
		}
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleaseTypesCommand", func() {
		var (
			command commands.ReleaseTypesCommand
		)

		BeforeEach(func() {
			responseStatusCode = http.StatusOK

			command = commands.ReleaseTypesCommand{}

			response = pivnet.ReleaseTypesResponse{
				ReleaseTypes: releaseTypes,
			}
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("%s/releases/release_types", apiPrefix)),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("lists all release types", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			var returnedReleaseTypes []string

			err = json.Unmarshal(outBuffer.Bytes(), &returnedReleaseTypes)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedReleaseTypes).To(Equal(releaseTypes))
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
})
