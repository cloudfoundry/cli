package util_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cloudfoundry/cli/cf/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONHelpers", func() {
	Describe("GetJSONFromFlagValue", func() {
		It("returns JSON bytes when given a file name prefixed with @", func() {
			tempfile, err := ioutil.TempFile("", "get-json-test")
			Expect(err).NotTo(HaveOccurred())
			jsonData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)

			jsonBytes, err := util.GetJSONFromFlagValue("@" + tempfile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).To(Equal([]byte(jsonData)))
		})

		It("returns an error when given an empty string", func() {
			_, err := util.GetJSONFromFlagValue("")
			Expect(err).To(HaveOccurred())
		})

		It("returns JSON bytes when given a file name not prefixed with @", func() {
			tempfile, err := ioutil.TempFile("", "get-json-test")
			Expect(err).NotTo(HaveOccurred())
			jsonData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)

			jsonBytes, err := util.GetJSONFromFlagValue(tempfile.Name())
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).To(Equal([]byte(jsonData)))
		})

		It("returns JSON bytes when given a file name not prefixed with @ and wrapped in double quotes", func() {
			tempfile, err := ioutil.TempFile("", "get-json-test")
			Expect(err).NotTo(HaveOccurred())
			jsonData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)

			jsonBytes, err := util.GetJSONFromFlagValue(`"` + tempfile.Name() + `"`)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).To(Equal([]byte(jsonData)))
		})

		It("returns JSON bytes when given a file name prefixed with @ and wrapped in double quotes after the @", func() {
			tempfile, err := ioutil.TempFile("", "get-json-test")
			Expect(err).NotTo(HaveOccurred())
			jsonData := `{"foo": "bar"}`
			ioutil.WriteFile(tempfile.Name(), []byte(jsonData), os.ModePerm)

			jsonBytes, err := util.GetJSONFromFlagValue(fmt.Sprintf(`@"%s"`, tempfile.Name()))
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).To(Equal([]byte(jsonData)))
		})

		It("returns JSON bytes when given literal JSON wrapped with single quotes", func() {
			jsonData := `'{"foo": "bar"}'`
			jsonBytes, err := util.GetJSONFromFlagValue(jsonData)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).To(Equal([]byte(strings.Trim(jsonData, `'`))))
		})

		It("returns JSON bytes when given literal JSON wrapped with double quotes", func() {
			jsonData := `"{"foo": "bar"}"`
			jsonBytes, err := util.GetJSONFromFlagValue(jsonData)
			Expect(err).NotTo(HaveOccurred())
			Expect(jsonBytes).To(Equal([]byte(strings.Trim(jsonData, `"`))))
		})

		It("returns an error when it cannot read the file prefixed with @", func() {
			_, err := util.GetJSONFromFlagValue("@nonexistent-file")
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when it cannot read the file not prefixed with @", func() {
			_, err := util.GetJSONFromFlagValue("nonexistent-file")
			Expect(err).To(HaveOccurred())
		})
	})
})
