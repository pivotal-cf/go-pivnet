package productfile_test

import (
	"bytes"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/productfile"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/productfile/productfilefakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler/errorhandlerfakes"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

var _ = Describe("productfile commands", func() {
	var (
		fakePivnetClient *productfilefakes.FakePivnetClient

		fakeErrorHandler *errorhandlerfakes.FakeErrorHandler

		outBuffer bytes.Buffer

		productfiles []pivnet.ProductFile

		client *productfile.ProductFileClient
	)

	BeforeEach(func() {
		fakePivnetClient = &productfilefakes.FakePivnetClient{}

		outBuffer = bytes.Buffer{}

		fakeErrorHandler = &errorhandlerfakes.FakeErrorHandler{}

		productfiles = []pivnet.ProductFile{
			{
				ID: 1234,
			},
			{
				ID: 2345,
			},
		}

		client = productfile.NewProductFileClient(
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

			fakePivnetClient.GetProductFilesReturns(productfiles, nil)
		})

		It("lists all ProductFiles", func() {
			err := client.List(productSlug, releaseVersion)
			Expect(err).NotTo(HaveOccurred())

			var returnedProductFiles []pivnet.ProductFile
			err = json.Unmarshal(outBuffer.Bytes(), &returnedProductFiles)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedProductFiles).To(Equal(productfiles))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("productfiles error")
				fakePivnetClient.GetProductFilesReturns(nil, expectedErr)
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
				fakePivnetClient.GetProductFilesForReleaseReturns(productfiles, nil)
			})

			It("lists all ProductFiles", func() {
				err := client.List(productSlug, releaseVersion)
				Expect(err).NotTo(HaveOccurred())

				var returnedProductFiles []pivnet.ProductFile
				err = json.Unmarshal(outBuffer.Bytes(), &returnedProductFiles)
				Expect(err).NotTo(HaveOccurred())

				Expect(returnedProductFiles).To(Equal(productfiles))
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
					expectedErr = errors.New("productfiles error")
					fakePivnetClient.GetProductFilesForReleaseReturns(nil, expectedErr)
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
			productSlug    string
			releaseVersion string
			productFileID  int
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			releaseVersion = ""
			productFileID = productfiles[0].ID

			fakePivnetClient.GetProductFileReturns(productfiles[0], nil)
		})

		It("gets ProductFile", func() {
			err := client.Get(productSlug, releaseVersion, productFileID)
			Expect(err).NotTo(HaveOccurred())

			var returnedProductFile pivnet.ProductFile
			err = json.Unmarshal(outBuffer.Bytes(), &returnedProductFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedProductFile).To(Equal(productfiles[0]))
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("productfile error")
				fakePivnetClient.GetProductFileReturns(pivnet.ProductFile{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.Get(productSlug, releaseVersion, productFileID)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})

		Context("when release version is not empty", func() {
			BeforeEach(func() {
				releaseVersion = "some-release-version"
				fakePivnetClient.GetProductFileForReleaseReturns(productfiles[0], nil)
			})

			It("gets ProductFile", func() {
				err := client.Get(productSlug, releaseVersion, productFileID)
				Expect(err).NotTo(HaveOccurred())

				var returnedProductFile pivnet.ProductFile
				err = json.Unmarshal(outBuffer.Bytes(), &returnedProductFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(returnedProductFile).To(Equal(productfiles[0]))
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
					err := client.Get(productSlug, releaseVersion, productFileID)
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
					expectedErr = errors.New("productfiles error")
					fakePivnetClient.GetProductFileForReleaseReturns(pivnet.ProductFile{}, expectedErr)
				})

				It("invokes the error handler", func() {
					err := client.Get(productSlug, releaseVersion, productFileID)
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
					Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
				})
			})
		})
	})

	Describe("AddToRelease", func() {
		var (
			productSlug    string
			releaseVersion string
			productFileID  int
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			releaseVersion = "release-version"
			productFileID = productfiles[0].ID

			fakePivnetClient.AddProductFileReturns(nil)
		})

		It("deletes ProductFile", func() {
			err := client.AddToRelease(productSlug, releaseVersion, productFileID)
			Expect(err).NotTo(HaveOccurred())
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
				err := client.AddToRelease(productSlug, releaseVersion, productFileID)
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
				expectedErr = errors.New("productfile error")
				fakePivnetClient.AddProductFileReturns(expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.AddToRelease(productSlug, releaseVersion, productFileID)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})

	Describe("RemoveFromRelease", func() {
		var (
			productSlug    string
			releaseVersion string
			productFileID  int
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			releaseVersion = "release-version"
			productFileID = productfiles[0].ID

			fakePivnetClient.RemoveProductFileReturns(nil)
		})

		It("deletes ProductFile", func() {
			err := client.RemoveFromRelease(productSlug, releaseVersion, productFileID)
			Expect(err).NotTo(HaveOccurred())
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
				err := client.RemoveFromRelease(productSlug, releaseVersion, productFileID)
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
				expectedErr = errors.New("productfile error")
				fakePivnetClient.RemoveProductFileReturns(expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.RemoveFromRelease(productSlug, releaseVersion, productFileID)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})

	Describe("Delete", func() {
		var (
			productSlug   string
			productFileID int
		)

		BeforeEach(func() {
			productSlug = "some-product-slug"
			productFileID = productfiles[0].ID

			fakePivnetClient.DeleteProductFileReturns(productfiles[0], nil)
		})

		It("deletes ProductFile", func() {
			err := client.Delete(productSlug, productFileID)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("productfile error")
				fakePivnetClient.DeleteProductFileReturns(pivnet.ProductFile{}, expectedErr)
			})

			It("invokes the error handler", func() {
				err := client.Delete(productSlug, productFileID)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeErrorHandler.HandleErrorCallCount()).To(Equal(1))
				Expect(fakeErrorHandler.HandleErrorArgsForCall(0)).To(Equal(expectedErr))
			})
		})
	})
})
