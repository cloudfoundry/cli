package manifestparser_test

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/util/manifestparser"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	var parser *Parser

	BeforeEach(func() {
		parser = NewParser()
	})

	Describe("NewParser", func() {
		It("returns a parser", func() {
			Expect(parser).ToNot(BeNil())
		})
	})

	Describe("Parse", func() {
		var (
			manifestPath string
			manifest     map[string]interface{}

			executeErr error
		)

		JustBeforeEach(func() {
			tmpfile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			manifestPath = tmpfile.Name()
			Expect(tmpfile.Close()).ToNot(HaveOccurred())

			WriteManifest(manifestPath, manifest)

			executeErr = parser.Parse(manifestPath)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(manifestPath)).ToNot(HaveOccurred())
		})

		Context("when given a valid manifest file", func() {
			BeforeEach(func() {
				manifest = map[string]interface{}{
					"applications": []map[string]string{
						{
							"name": "app-1",
						},
						{
							"name": "app-2",
						},
					},
				}
			})

			It("returns nil and sets the applications", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})

		Context("when given an invalid manifest file", func() {
			BeforeEach(func() {
				manifest = map[string]interface{}{}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("must have at least one application"))
			})
		})
	})

	Describe("AppNames", func() {
		Context("when given a valid manifest file", func() {
			BeforeEach(func() {
				parser.Applications = []Application{{Name: "app-1"}, {Name: "app-2"}}
			})

			It("gets the app names", func() {
				appNames := parser.AppNames()
				Expect(appNames).To(ConsistOf("app-1", "app-2"))
			})
		})
	})

	Describe("RawManifest", func() {
		Context("when given an app name", func() {
			Context("when app is successfully marshalled", func() {
				var manifestPath string

				BeforeEach(func() {
					tmpfile, err := ioutil.TempFile("", "")
					Expect(err).ToNot(HaveOccurred())
					manifestPath = tmpfile.Name()
					Expect(tmpfile.Close()).ToNot(HaveOccurred())

					manifest := map[string]interface{}{
						"applications": []map[string]string{
							{
								"name": "app-1",
							},
							{
								"name": "app-2",
							},
						},
					}
					WriteManifest(manifestPath, manifest)

					executeErr := parser.Parse(manifestPath)
					Expect(executeErr).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(manifestPath)).ToNot(HaveOccurred())
				})

				It("gets the app's manifest", func() {
					rawManifest, err := parser.RawManifest("app-1")
					Expect(err).ToNot(HaveOccurred())
					Expect(rawManifest).To(MatchYAML(`---
applications:
- name: app-1

- name: app-2
`))

					rawManifest, err = parser.RawManifest("app-2")
					Expect(err).ToNot(HaveOccurred())
					Expect(rawManifest).To(MatchYAML(`---
applications:
- name: app-1

- name: app-2
`))
				})
			})

			PContext("when app marshalling errors", func() {
				It("returns an error", func() {})
			})
		})
	})
})
