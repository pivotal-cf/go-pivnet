package commands_test

import (
	"errors"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/commandsfakes"
)

var _ = Describe("file group commands", func() {
	var (
		field reflect.StructField

		fakeFileGroupClient *commandsfakes.FakeFileGroupClient
	)

	BeforeEach(func() {
		fakeFileGroupClient = &commandsfakes.FakeFileGroupClient{}

		commands.NewFileGroupClient = func() commands.FileGroupClient {
			return fakeFileGroupClient
		}
	})

	Describe("FileGroupsCommand", func() {
		var (
			cmd *commands.FileGroupsCommand
		)

		BeforeEach(func() {
			cmd = &commands.FileGroupsCommand{}
		})

		It("invokes the FileGroup client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileGroupClient.ListCallCount()).To(Equal(1))
		})

		Context("when the FileGroup client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeFileGroupClient.ListReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupsCommand{}, "ProductSlug")
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
				field = fieldFor(commands.FileGroupsCommand{}, "ReleaseVersion")
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

	Describe("FileGroupCommand", func() {
		var (
			cmd *commands.FileGroupCommand
		)

		BeforeEach(func() {
			cmd = &commands.FileGroupCommand{}
		})

		It("invokes the FileGroup client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileGroupClient.GetCallCount()).To(Equal(1))
		})

		Context("when the FileGroup client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeFileGroupClient.GetReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupCommand{}, "ProductSlug")
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

		Describe("FileGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.FileGroupCommand{}, "FileGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("file-group-id"))
			})
		})
	})

	Describe("CreateFileGroupCommand", func() {
		var (
			cmd *commands.CreateFileGroupCommand
		)

		BeforeEach(func() {
			cmd = &commands.CreateFileGroupCommand{}
		})

		It("invokes the FileGroup client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileGroupClient.CreateCallCount()).To(Equal(1))
		})

		Context("when the FileGroup client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeFileGroupClient.CreateReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.CreateFileGroupCommand{}, "ProductSlug")
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

		Describe("Name flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.CreateFileGroupCommand{}, "Name")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("name"))
			})
		})
	})

	Describe("UpdateFileGroupCommand", func() {
		var (
			cmd commands.UpdateFileGroupCommand
		)

		BeforeEach(func() {
			cmd = commands.UpdateFileGroupCommand{}
		})

		It("invokes the FileGroup client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileGroupClient.UpdateCallCount()).To(Equal(1))
		})

		Context("when the FileGroup client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeFileGroupClient.UpdateReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})
	})

	Describe("DeleteFileGroupCommand", func() {
		var (
			cmd *commands.DeleteFileGroupCommand
		)

		BeforeEach(func() {
			cmd = &commands.DeleteFileGroupCommand{}
		})

		It("invokes the FileGroup client", func() {
			err := cmd.Execute(nil)

			Expect(err).NotTo(HaveOccurred())

			Expect(fakeFileGroupClient.DeleteCallCount()).To(Equal(1))
		})

		Context("when the FileGroup client returns an error", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				expectedErr = errors.New("expected error")
				fakeFileGroupClient.DeleteReturns(expectedErr)
			})

			It("forwards the error", func() {
				err := cmd.Execute(nil)

				Expect(err).To(Equal(expectedErr))
			})
		})

		Describe("ProductSlug flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteFileGroupCommand{}, "ProductSlug")
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

		Describe("FileGroupID flag", func() {
			BeforeEach(func() {
				field = fieldFor(commands.DeleteFileGroupCommand{}, "FileGroupID")
			})

			It("is required", func() {
				Expect(isRequired(field)).To(BeTrue())
			})

			It("contains long name", func() {
				Expect(longTag(field)).To(Equal("file-group-id"))
			})
		})
	})
})
