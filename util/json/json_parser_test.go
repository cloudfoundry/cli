package json_test

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/util/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON Parser", func() {
	Describe("ParseJSONArray", func() {
		var filename string
		var tmpFile *os.File

		Context("when everything is proper", func() {
			BeforeEach(func() {
				tmpFile, _ = ioutil.TempFile("", "WONDERFULFILEWHOSENAMEISHARDTOREADBUTCONTAINSVALIDJSON")
				filename = tmpFile.Name()
				ioutil.WriteFile(filename, []byte("[{\"akey\": \"avalue\"}]"), 0644)
			})

			AfterEach(func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			})

			It("converts a json file into an unmarshalled slice of string->string map objects", func() {
				stringMaps, err := json.ParseJSONArray(filename)
				Expect(err).To(BeNil())
				Expect(stringMaps[0]["akey"]).To(Equal("avalue"))
			})
		})

		Context("when the JSON is invalid", func() {
			BeforeEach(func() {
				tmpFile, _ = ioutil.TempFile("", "TERRIBLEFILECONTAININGINVALIDJSONWHICHMAKESEVERYTHINGTERRIBLEANDSTILLHASANAMETHATSHARDTOREAD")
				filename = tmpFile.Name()
				ioutil.WriteFile(filename, []byte("SCARY NOISES}"), 0644)
			})

			AfterEach(func() {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
			})

			It("tries to convert the json file but fails because it was given something it didn't like", func() {
				_, err := json.ParseJSONArray(filename)
				Expect(err).To(MatchError("Incorrect json format: invalid character 'S' looking for beginning of value"))
			})
		})
	})

	Describe("ParseJSONFromFileOrString", func() {
		Context("when the input is empty", func() {
			It("returns nil", func() {
				result, err := json.ParseJSONFromFileOrString("")

				Expect(result).To(BeNil())
				Expect(err).To(BeNil())
			})
		})

		Context("when the input is a file", func() {
			var jsonFile *os.File
			var fileContent string

			AfterEach(func() {
				if jsonFile != nil {
					jsonFile.Close()
					os.Remove(jsonFile.Name())
				}
			})

			BeforeEach(func() {
				fileContent = `{"foo": "bar"}`
			})

			JustBeforeEach(func() {
				var err error
				jsonFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				err = ioutil.WriteFile(jsonFile.Name(), []byte(fileContent), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the parsed json from the file", func() {
				result, err := json.ParseJSONFromFileOrString(jsonFile.Name())
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("when the file contains invalid json", func() {
				BeforeEach(func() {
					fileContent = `badtimes`
				})

				It("returns an error", func() {
					_, err := json.ParseJSONFromFileOrString(jsonFile.Name())
					Expect(err).To(MatchError("Incorrect json format: invalid character 'b' looking for beginning of value"))
				})
			})
		})

		Context("when the input is considered a json string (when it is not a file path)", func() {
			var jsonString string

			BeforeEach(func() {
				jsonString = `{"foo": "bar"}`
			})

			It("returns the parsed json", func() {
				result, err := json.ParseJSONFromFileOrString(jsonString)
				Expect(err).NotTo(HaveOccurred())

				Expect(result).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("when the JSON is invalid", func() {
				BeforeEach(func() {
					jsonString = "SOMETHING IS WRONG"
				})

				It("returns a json parse error", func() {
					_, err := json.ParseJSONFromFileOrString(jsonString)
					Expect(err).To(MatchError("Incorrect json format: invalid character 'S' looking for beginning of value"))
				})
			})
		})

		Context("when the input is neither a file nor a json string", func() {
			var invalidInput string

			BeforeEach(func() {
				invalidInput = "boo"
			})

			It("returns an error", func() {
				_, err := json.ParseJSONFromFileOrString(invalidInput)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
