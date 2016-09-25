package commands_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/commandsfakes"
)

var _ = Describe("product file commands", func() {
	var (
		field reflect.StructField

		fakeProductFileClient *commandsfakes.FakeProductFileClient
	)

	BeforeEach(func() {
		fakeProductFileClient = &commandsfakes.FakeProductFileClient{}

		commands.NewProductFileClient = func() commands.ProductFileClient {
			return fakeProductFileClient
		}
	})

	Describe("ProductFilesCommand", func() {
		var (
			cmd commands.ProductFilesCommand
		)

		BeforeEach(func() {
			cmd = commands.ProductFilesCommand{}
		})

		It("invokes the ProductFile client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeProductFileClient.ListCallCount()).To(Equal(1))
		})

		Context("when the ProductFile client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeProductFileClient.ListReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFilesCommand{}, "ProductSlug")
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

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFilesCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("r"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})
	})

	Describe("ProductFileCommand", func() {
		var (
			cmd commands.ProductFileCommand
		)

		BeforeEach(func() {
			cmd = commands.ProductFileCommand{}
		})

		It("invokes the ProductFile client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeProductFileClient.GetCallCount()).To(Equal(1))
		})

		Context("when the ProductFile client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeProductFileClient.GetReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFileCommand{}, "ProductSlug")
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

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFileCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("r"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.ProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("i"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})
	})

	Describe("AddProductFileCommand", func() {
		var (
			cmd commands.AddProductFileCommand

			releaseVersion string
			fileGroupID    int
		)

		BeforeEach(func() {
			releaseVersion = "some release version"
			fileGroupID = 5432

			cmd = commands.AddProductFileCommand{ReleaseVersion: &releaseVersion}
		})

		Context("when neither releaseVersion nor fileGroupID are provided", func() {
			BeforeEach(func() {
				cmd = commands.AddProductFileCommand{}
			})

			It("returns an error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("one of release-version or file-group-id"))
			})
		})

		Context("when both releaseVersion and fileGroupID are provided", func() {
			BeforeEach(func() {
				cmd = commands.AddProductFileCommand{
					ReleaseVersion: &releaseVersion,
					FileGroupID:    &fileGroupID,
				}
			})

			It("returns an error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only one of release-version or file-group-id"))
			})
		})

		Context("when release-version is provided", func() {
			BeforeEach(func() {
				cmd = commands.AddProductFileCommand{ReleaseVersion: &releaseVersion}
			})

			It("invokes the ProductFile client", func() {
				err := cmd.Execute(nil)

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeProductFileClient.AddToReleaseCallCount()).To(Equal(1))
			})

			Context("when the ProductFile client returns an error", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("expected error")
					fakeProductFileClient.AddToReleaseReturns(expectedErr)
				})

				It("forwards the error", func() {
					err := cmd.Execute(nil)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})

		Context("when file-group-id is provided", func() {
			BeforeEach(func() {
				cmd = commands.AddProductFileCommand{FileGroupID: &fileGroupID}
			})

			It("invokes the ProductFile client", func() {
				err := cmd.Execute(nil)

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeProductFileClient.AddToFileGroupCallCount()).To(Equal(1))
			})

			Context("when the ProductFile client returns an error", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("expected error")
					fakeProductFileClient.AddToFileGroupReturns(expectedErr)
				})

				It("forwards the error", func() {
					err := cmd.Execute(nil)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "ProductSlug")
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

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("i"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("r"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})

		Describe("FileGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.AddProductFileCommand{}, "FileGroupID")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("f"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("file-group-id"))
			})
		})
	})

	Describe("RemoveProductFileCommand", func() {
		var (
			cmd commands.RemoveProductFileCommand

			releaseVersion string
			fileGroupID    int
		)

		BeforeEach(func() {
			releaseVersion = "some release version"
			fileGroupID = 5432

			cmd = commands.RemoveProductFileCommand{ReleaseVersion: &releaseVersion}
		})

		Context("when neither releaseVersion nor fileGroupID are provided", func() {
			BeforeEach(func() {
				cmd = commands.RemoveProductFileCommand{}
			})

			It("returns an error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("one of release-version or file-group-id"))
			})
		})

		Context("when both releaseVersion and fileGroupID are provided", func() {
			BeforeEach(func() {
				cmd = commands.RemoveProductFileCommand{
					ReleaseVersion: &releaseVersion,
					FileGroupID:    &fileGroupID,
				}
			})

			It("returns an error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only one of release-version or file-group-id"))
			})
		})

		Context("when release-version is provided", func() {
			BeforeEach(func() {
				cmd = commands.RemoveProductFileCommand{ReleaseVersion: &releaseVersion}
			})

			It("invokes the ProductFile client", func() {
				err := cmd.Execute(nil)

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeProductFileClient.RemoveFromReleaseCallCount()).To(Equal(1))
			})

			Context("when the ProductFile client returns an error", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("expected error")
					fakeProductFileClient.RemoveFromReleaseReturns(expectedErr)
				})

				It("forwards the error", func() {
					err := cmd.Execute(nil)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})

		Context("when file-group-id is provided", func() {
			BeforeEach(func() {
				cmd = commands.RemoveProductFileCommand{FileGroupID: &fileGroupID}
			})

			It("invokes the ProductFile client", func() {
				err := cmd.Execute(nil)

				Expect(err).NotTo(HaveOccurred())

				Expect(fakeProductFileClient.RemoveFromFileGroupCallCount()).To(Equal(1))
			})

			Context("when the ProductFile client returns an error", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("expected error")
					fakeProductFileClient.RemoveFromFileGroupReturns(expectedErr)
				})

				It("forwards the error", func() {
					err := cmd.Execute(nil)

					Expect(err).To(Equal(expectedErr))
				})
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "ProductSlug")
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

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("i"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "ReleaseVersion")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("r"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})
		})

		Describe("FileGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.RemoveProductFileCommand{}, "FileGroupID")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("file-group-id"))
			})
		})
	})

	Describe("DeleteProductFileCommand", func() {
		var (
			cmd commands.DeleteProductFileCommand
		)

		BeforeEach(func() {
			cmd = commands.DeleteProductFileCommand{}
		})

		It("invokes the ProductFile client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeProductFileClient.DeleteCallCount()).To(Equal(1))
		})

		Context("when the ProductFile client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeProductFileClient.DeleteReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteProductFileCommand{}, "ProductSlug")
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

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("i"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})
	})

	Describe("DownloadProductFileCommand", func() {
		var (
			cmd commands.DownloadProductFileCommand
		)

		BeforeEach(func() {
			cmd = commands.DownloadProductFileCommand{}
		})

		It("invokes the ProductFile client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeProductFileClient.DownloadCallCount()).To(Equal(1))
		})

		Context("when the ProductFile client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeProductFileClient.DownloadReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "ProductSlug")
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

		Describe("ReleaseVersion flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "ReleaseVersion")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("release-version"))
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("r"))
			})
		})

		Describe("ProductFileID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "ProductFileID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains short name", func() {
				Expect(shortTag(field)).To(Equal("i"))
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("product-file-id"))
			})
		})

		Describe("Filepath flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "Filepath")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("filepath"))
			})
		})

		Describe("AcceptEULA flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DownloadProductFileCommand{}, "AcceptEULA")
			})

			It("is not required", func() {
				Expect(isRequired(field)).To(BeFalse())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("accept-eula"))
			})
		})
	})
})
