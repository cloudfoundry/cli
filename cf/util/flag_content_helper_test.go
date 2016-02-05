package util_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/cli/cf/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONHelpers", func() {
	Describe("GetContentsFromFlagValue", func() {
		It("returns bytes when given a file name prefixed with @", func() {
			tempfile, err := ioutil.TempFile("", "get-data-test")
			Expect(err).NotTo(HaveOccurred())
			fileData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(fileData), os.ModePerm)

			bs, err := util.GetContentsFromFlagValue("@" + tempfile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte(fileData)))
		})

		It("returns an error when given an empty string", func() {
			_, err := util.GetContentsFromFlagValue("")
			Expect(err).To(HaveOccurred())
		})

		It("returns bytes when given a file name not prefixed with @", func() {
			tempfile, err := ioutil.TempFile("", "get-data-test")
			Expect(err).NotTo(HaveOccurred())
			fileData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(fileData), os.ModePerm)

			bs, err := util.GetContentsFromFlagValue(tempfile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte(fileData)))
		})

		It("returns bytes when given a file name not prefixed with @ and wrapped in double quotes", func() {
			tempfile, err := ioutil.TempFile("", "get-data-test")
			Expect(err).NotTo(HaveOccurred())
			fileData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(fileData), os.ModePerm)

			bs, err := util.GetContentsFromFlagValue(`"` + tempfile.Name() + `"`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte(fileData)))
		})

		It("returns bytes when given a file name prefixed with @ and wrapped in double quotes after the @", func() {
			tempfile, err := ioutil.TempFile("", "get-data-test")
			Expect(err).NotTo(HaveOccurred())
			fileData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(fileData), os.ModePerm)

			bs, err := util.GetContentsFromFlagValue(fmt.Sprintf(`@"%s"`, tempfile.Name()))
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte(fileData)))
		})

		It("returns bytes when given something that isn't a file wrapped with single quotes", func() {
			bs, err := util.GetContentsFromFlagValue(`'param1=value1&param2=value2'`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("param1=value1&param2=value2")))
		})

		It("returns bytes when given something that isn't a file wrapped with double quotes", func() {
			bs, err := util.GetContentsFromFlagValue(`"param1=value1&param2=value2"`)
			Expect(err).NotTo(HaveOccurred())
			Expect(bs).To(Equal([]byte("param1=value1&param2=value2")))
		})

		It("returns an error when it cannot read the file prefixed with @", func() {
			_, err := util.GetContentsFromFlagValue("@nonexistent-file")
			Expect(err).To(HaveOccurred())
		})
	})
})
