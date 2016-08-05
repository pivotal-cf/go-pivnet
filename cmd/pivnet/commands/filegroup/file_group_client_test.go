package filegroup_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/filegroup"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/filegroup/filegroupfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var _ = Describe("filegroup commands", func() {
	var (
		fakePivnetClient *filegroupfakes.FakePivnetClient

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		outBuffer bytes.Buffer

		filegroups []pivnet.FileGroup

		client *filegroup.FileGroupClient
	)

	BeforeEach(func() {
		fakePivnetClient = &filegroupfakes.FakePivnetClient{}

		outBuffer = bytes.Buffer{}

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}

		filegroups = []pivnet.FileGroup{
			{
				ID: 1234,
			},
			{
				ID: 2345,
			},
		}

		client = filegroup.NewFileGroupClient(
			fakePivnetClient,
			fakeErrorHandler,
			printer.PrintAsJSON,
			&outBuffer,
			printer.NewPrinter(&outBuffer),
		)
	})

	Describe("List", func() {
		var (
			productSlug    string
			releaseVersion string
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			releaseVersion = ""

			fakePivnetClient.FileGroupsReturns(filegroups, nil)
		})

		It("lists all FileGroups", func() {
			err := client.List(productSlug, releaseVersion)
			Expect(err).NotTo(HaveOccurred())

			var returnedFileGroups []pivnet.FileGroup
			err = json.Unmarshal(outBuffer.Bytes(), &returnedFileGroups)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedFileGroups).To(Equal(filegroups))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("filegroups error")
				fakePivnetClient.FileGroupsReturns(nil, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.List(productSlug, releaseVersion)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})

		Context("when release version is not empty", func() {
			BeforeEach(func() {
				releaseVersion = "some-release-version"
				fakePivnetClient.FileGroupsForReleaseReturns(filegroups, nil)
			})

			Context("when there is an error getting release", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("releases error")
					fakePivnetClient.ReleaseForProductVersionReturns(pivnet.Release{}, expectedErr)
				})

				It("invokes the error handler", func() {
					err := client.List(productSlug, releaseVersion)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
					Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
				})
			})

			Context("when there is an error", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("filegroups error")
					fakePivnetClient.FileGroupsForReleaseReturns(nil, expectedErr)
				})

				It("invokes the error handler", func() {
					err := client.List(productSlug, releaseVersion)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
					Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
				})
			})
		})
	})

	Describe("Get", func() {
		var (
			productSlug string
			fileGroupID int
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			fileGroupID = filegroups[0].ID

			fakePivnetClient.FileGroupReturns(filegroups[0], nil)
		})

		It("gets FileGroup", func() {
			err := client.Get(productSlug, fileGroupID)
			Expect(err).NotTo(HaveOccurred())

			var returnedFileGroup pivnet.FileGroup
			err = json.Unmarshal(outBuffer.Bytes(), &returnedFileGroup)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedFileGroup).To(Equal(filegroups[0]))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("filegroup error")
				fakePivnetClient.FileGroupReturns(pivnet.FileGroup{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.Get(productSlug, fileGroupID)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})

	Describe("Delete", func() {
		var (
			productSlug string
			fileGroupID int
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			fileGroupID = filegroups[0].ID

			fakePivnetClient.DeleteFileGroupReturns(filegroups[0], nil)
		})

		It("deletes FileGroup", func() {
			err := client.Delete(productSlug, fileGroupID)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("filegroup error")
				fakePivnetClient.DeleteFileGroupReturns(pivnet.FileGroup{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.Delete(productSlug, fileGroupID)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})
})
