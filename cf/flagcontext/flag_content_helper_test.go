package flagcontext_test

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cli/cf/flagcontext"

	"code.cloudfoundry.org/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flag Content Helpers", func() {
	Describe("GetContentsFromOptionalFlagValue", func() {
		It("returns an empty byte slice when given an empty string", func() {
			bs, err := flagcontext.GetContentsFromOptionalFlagValue("")
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte{}))
		})

		It("returns bytes when given a file name prefixed with @", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromOptionalFlagValue("@" + tmpFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given a file name not prefixed with @", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromOptionalFlagValue(tmpFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given a file name not prefixed with @ and wrapped in double quotes", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromOptionalFlagValue(fmt.Sprintf(`"%s"`, tmpFile.Name()))
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given a file name prefixed with @ and wrapped in double quotes after the @", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromOptionalFlagValue(fmt.Sprintf(`@"%s"`, tmpFile.Name()))
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given something that isn't a file wrapped with single quotes", func() {
			bs, err := flagcontext.GetContentsFromOptionalFlagValue(`'param1=value1&param2=value2'`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("param1=value1&param2=value2")))
		})

		It("returns bytes when given something that isn't a file wrapped with double quotes", func() {
			bs, err := flagcontext.GetContentsFromOptionalFlagValue(`"param1=value1&param2=value2"`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("param1=value1&param2=value2")))
		})

		It("returns an error when it cannot read the file prefixed with @", func() {
			_, err := flagcontext.GetContentsFromOptionalFlagValue("@nonexistent-file")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetContentsFromFlagValue", func() {
		It("returns an error when given an empty string", func() {
			_, err := flagcontext.GetContentsFromFlagValue("")
			Expect(err).To(HaveOccurred())
		})

		It("returns bytes when given a file name prefixed with @", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromFlagValue("@" + tmpFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given a file name not prefixed with @", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromFlagValue(tmpFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given a file name not prefixed with @ and wrapped in double quotes", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromFlagValue(fmt.Sprintf(`"%s"`, tmpFile.Name()))
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given a file name prefixed with @ and wrapped in double quotes after the @", func() {
			fileutils.TempFile("get-data-test", func(tmpFile *os.File, err error) {
				Expect(err).NotTo(HaveOccurred())
				fileData := `{"foo": "bar"}`
				tmpFile.WriteString(fileData)

				bs, err := flagcontext.GetContentsFromFlagValue(fmt.Sprintf(`@"%s"`, tmpFile.Name()))
				Expect(err).NotTo(HaveOccurred())
				Expect(bs).To(Equal([]byte(fileData)))
			})
		})

		It("returns bytes when given something that isn't a file wrapped with single quotes", func() {
			bs, err := flagcontext.GetContentsFromFlagValue(`'param1=value1&param2=value2'`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("param1=value1&param2=value2")))
		})

		It("returns bytes when given something that isn't a file wrapped with double quotes", func() {
			bs, err := flagcontext.GetContentsFromFlagValue(`"param1=value1&param2=value2"`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("param1=value1&param2=value2")))
		})

		It("returns an error when it cannot read the file prefixed with @", func() {
			_, err := flagcontext.GetContentsFromFlagValue("@nonexistent-file")
			Expect(err).To(HaveOccurred())
		})
	})
})
