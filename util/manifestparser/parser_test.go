package manifestparser_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/manifestparser"

	"github.com/cloudfoundry/bosh-cli/director/template"
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

	Describe("AppNames", func() {
		When("given a valid manifest file", func() {
			BeforeEach(func() {
				parser.Applications = []Application{
					{ApplicationModel: ApplicationModel{Name: "app-1"}, FullUnmarshalledApplication: nil},
					{ApplicationModel: ApplicationModel{Name: "app-2"}, FullUnmarshalledApplication: nil}}
			})

			It("gets the app names", func() {
				appNames := parser.AppNames()
				Expect(appNames).To(ConsistOf("app-1", "app-2"))
			})
		})
	})

	Describe("ApplyNoRouteOverride", func() {
		var (
			executeErr error
			appName    string
		)

		BeforeEach(func() {
			parser.Applications = []Application{{ApplicationModel{Name: "appName", NoRoute: false}, nil}}
		})

		JustBeforeEach(func() {
			executeErr = parser.ApplyNoRouteOverride(appName, true)
		})

		When("App is in the manifest", func() {
			It("returns nil", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(parser.Applications[0].NoRoute).To(BeTrue())
			})
		})
	})
	Describe("ContainsManifest", func() {
		var (
			pathToManifest string
		)

		BeforeEach(func() {
			tempFile, err := ioutil.TempFile("", "contains-manifest-test")
			Expect(err).ToNot(HaveOccurred())
			pathToManifest = tempFile.Name()
			Expect(tempFile.Close()).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(pathToManifest)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the manifest is parsed successfully", func() {
			BeforeEach(func() {
				rawManifest := []byte(`---
applications:
- name: spark
`)
				err := ioutil.WriteFile(pathToManifest, rawManifest, 0666)
				Expect(err).ToNot(HaveOccurred())

				err = parser.InterpolateAndParse(pathToManifest, nil, nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns true", func() {
				Expect(parser.ContainsManifest()).To(BeTrue())
			})
		})

		Context("when the manifest is not parsed successfully", func() {
			BeforeEach(func() {
				rawManifest := []byte(`---
applications:
`)
				err := ioutil.WriteFile(pathToManifest, rawManifest, 0666)
				Expect(err).ToNot(HaveOccurred())

				err = parser.InterpolateAndParse(pathToManifest, nil, nil)
				Expect(err).To(HaveOccurred())
			})

			It("returns false", func() {
				Expect(parser.ContainsManifest()).To(BeFalse())
			})
		})

		Context("when the manifest has not been parsed", func() {
			It("returns false", func() {
				Expect(parser.ContainsManifest()).To(BeFalse())
			})
		})
	})

	Describe("ContainsMultipleApps", func() {
		When("given a valid manifest file with multiple apps", func() {
			BeforeEach(func() {
				parser.Applications = []Application{
					{ApplicationModel: ApplicationModel{Name: "app-1"}, FullUnmarshalledApplication: nil},
					{ApplicationModel: ApplicationModel{Name: "app-2"}, FullUnmarshalledApplication: nil}}
			})

			It("returns true", func() {
				Expect(parser.ContainsMultipleApps()).To(BeTrue())
			})
		})

		When("given a valid manifest file with a single app", func() {
			BeforeEach(func() {
				parser.Applications = []Application{{ApplicationModel: ApplicationModel{Name: "app-1"}, FullUnmarshalledApplication: nil}}
			})

			It("returns false", func() {
				Expect(parser.ContainsMultipleApps()).To(BeFalse())
			})
		})
	})

	Describe("ContainsPrivateDockerImages", func() {
		When("the manifest contains a docker image", func() {
			When("the image is public", func() {
				BeforeEach(func() {
					parser.Applications = []Application{
						{ApplicationModel: ApplicationModel{Name: "app-1", Docker: &Docker{Image: "image-1"}}, FullUnmarshalledApplication: nil},
						{ApplicationModel: ApplicationModel{Name: "app-2", Docker: &Docker{Image: "image-2"}}, FullUnmarshalledApplication: nil}}
				})

				It("returns false", func() {
					Expect(parser.ContainsPrivateDockerImages()).To(BeFalse())
				})
			})

			When("the image is private", func() {
				BeforeEach(func() {
					parser.Applications = []Application{
						{ApplicationModel: ApplicationModel{Name: "app-1", Docker: &Docker{Image: "image-1"}}},
						{ApplicationModel: ApplicationModel{Name: "app-2", Docker: &Docker{Image: "image-2", Username: "user"}}},
					}
				})

				It("returns true", func() {
					Expect(parser.ContainsPrivateDockerImages()).To(BeTrue())
				})
			})
		})

		When("the manifest does not contain a docker image", func() {
			BeforeEach(func() {
				parser.Applications = []Application{
					{ApplicationModel: ApplicationModel{Name: "app-1"}},
					{ApplicationModel: ApplicationModel{Name: "app-2"}},
				}
			})

			It("returns false", func() {
				Expect(parser.ContainsPrivateDockerImages()).To(BeFalse())
			})
		})
	})

	Describe("InterpolateAndParse", func() {
		var (
			pathToManifest   string
			pathsToVarsFiles []string
			vars             []template.VarKV

			executeErr error

			rawManifest []byte
		)

		BeforeEach(func() {
			tempFile, err := ioutil.TempFile("", "manifest-test-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).ToNot(HaveOccurred())
			pathToManifest = tempFile.Name()
			vars = nil

			pathsToVarsFiles = nil
		})

		AfterEach(func() {
			Expect(os.RemoveAll(pathToManifest)).ToNot(HaveOccurred())
			for _, path := range pathsToVarsFiles {
				Expect(os.RemoveAll(path)).ToNot(HaveOccurred())
			}
		})

		JustBeforeEach(func() {
			executeErr = parser.InterpolateAndParse(pathToManifest, pathsToVarsFiles, vars)
		})

		Context("regardless of whether the manifest needs interpolation", func() {
			BeforeEach(func() {
				rawManifest = []byte(`---
applications:
- name: spark
  memory: 1G
  instances: 2
- name: flame
  memory: 1G
  instances: 2
`)
				err := ioutil.WriteFile(pathToManifest, rawManifest, 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("parses the manifest properly", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(parser.AppNames()).To(ConsistOf("spark", "flame"))
				Expect(parser.PathToManifest).To(Equal(pathToManifest))
				Expect(parser.FullRawManifest()).To(MatchYAML(rawManifest))
			})
		})

		Context("invalid yaml is passed", func() {
			BeforeEach(func() {
				rawManifest = []byte("\t\t")
				err := ioutil.WriteFile(pathToManifest, rawManifest, 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("parses the manifest properly", func() {
				Expect(executeErr).To(HaveOccurred())
			})
		})

		When("the manifest contains variables that need interpolation", func() {
			BeforeEach(func() {
				rawManifest = []byte(`---
applications:
- name: ((var1))
- name: ((var2))
`)
				err := ioutil.WriteFile(pathToManifest, rawManifest, 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			When("only vars files are provided", func() {
				var (
					varsDir string
				)

				BeforeEach(func() {
					var err error
					varsDir, err = ioutil.TempDir("", "vars-test")
					Expect(err).ToNot(HaveOccurred())

					varsFilePath1 := filepath.Join(varsDir, "vars-1")
					err = ioutil.WriteFile(varsFilePath1, []byte("var1: spark"), 0666)
					Expect(err).ToNot(HaveOccurred())

					varsFilePath2 := filepath.Join(varsDir, "vars-2")
					err = ioutil.WriteFile(varsFilePath2, []byte("var2: flame"), 0666)
					Expect(err).ToNot(HaveOccurred())

					pathsToVarsFiles = append(pathsToVarsFiles, varsFilePath1, varsFilePath2)
				})

				AfterEach(func() {
					Expect(os.RemoveAll(varsDir)).ToNot(HaveOccurred())
				})

				When("multiple values for the same variable(s) are provided", func() {
					BeforeEach(func() {
						varsFilePath1 := filepath.Join(varsDir, "vars-1")
						err := ioutil.WriteFile(varsFilePath1, []byte("var1: garbageapp\nvar1: spark\nvar2: doesn't matter"), 0666)
						Expect(err).ToNot(HaveOccurred())

						varsFilePath2 := filepath.Join(varsDir, "vars-2")
						err = ioutil.WriteFile(varsFilePath2, []byte("var2: flame"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = append(pathsToVarsFiles, varsFilePath1, varsFilePath2)
					})

					It("interpolates the placeholder values", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(parser.AppNames()).To(ConsistOf("spark", "flame"))
					})
				})

				When("the provided files exists and contain valid yaml", func() {
					It("interpolates the placeholder values", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(parser.AppNames()).To(ConsistOf("spark", "flame"))
					})
				})

				When("a variable in the manifest is not provided in the vars file", func() {
					BeforeEach(func() {
						varsFilePath := filepath.Join(varsDir, "vars-1")
						err := ioutil.WriteFile(varsFilePath, []byte("notvar: foo"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = []string{varsFilePath}
					})

					It("returns an error", func() {
						Expect(executeErr.Error()).To(Equal("Expected to find variables: var1, var2"))
					})
				})

				When("the provided file path does not exist", func() {
					BeforeEach(func() {
						pathsToVarsFiles = []string{"garbagepath"}
					})

					It("returns an error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(os.IsNotExist(executeErr)).To(BeTrue())
					})
				})

				When("the provided file is not a valid yaml file", func() {
					BeforeEach(func() {
						varsFilePath := filepath.Join(varsDir, "vars-1")
						err := ioutil.WriteFile(varsFilePath, []byte(": bad"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = []string{varsFilePath}
					})

					It("returns an error", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(executeErr).To(MatchError(InvalidYAMLError{
							Err: errors.New("yaml: did not find expected key"),
						}))
					})
				})
			})

			When("only vars are provided", func() {
				BeforeEach(func() {
					vars = []template.VarKV{
						{Name: "var1", Value: "spark"},
						{Name: "var2", Value: "flame"},
					}
				})

				It("interpolates the placeholder values", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(parser.AppNames()).To(ConsistOf("spark", "flame"))
				})
			})

			When("vars and vars files are provided", func() {
				var varsFilePath string
				BeforeEach(func() {
					tmp, err := ioutil.TempFile("", "util-manifest-varsilfe")
					Expect(err).NotTo(HaveOccurred())
					Expect(tmp.Close()).NotTo(HaveOccurred())

					varsFilePath = tmp.Name()
					err = ioutil.WriteFile(varsFilePath, []byte("var1: spark\nvar2: 12345"), 0666)
					Expect(err).ToNot(HaveOccurred())

					pathsToVarsFiles = []string{varsFilePath}
					vars = []template.VarKV{
						{Name: "var2", Value: "flame"},
					}
				})

				AfterEach(func() {
					Expect(os.RemoveAll(varsFilePath)).ToNot(HaveOccurred())
				})

				It("interpolates the placeholder values, prioritizing the vars flag", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(parser.AppNames()).To(ConsistOf("spark", "flame"))
				})
			})
		})
	})

	Describe("RawAppManifest", func() {
		var (
			rawAppManifest []byte
			appName        string
			executeErr     error
			rawManifest    []byte
			pathToManifest string
			tmpMyPath      string
		)

		BeforeEach(func() {
			var err error

			appName = "spark"

			tmpMyPath, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			rawManifest = []byte(fmt.Sprintf(`---
applications:
- name: spark
  memory: 1G
  instances: 2
  docker:
    username: experiment
  path: %s
- name: flame
  memory: 1G
  instances: 2
  docker:
    username: experiment
`, tmpMyPath))

		})

		JustBeforeEach(func() {
			tempFile, err := ioutil.TempFile("", "manifest-test-")
			Expect(err).ToNot(HaveOccurred())
			Expect(tempFile.Close()).ToNot(HaveOccurred())
			pathToManifest = tempFile.Name()
			err = ioutil.WriteFile(pathToManifest, rawManifest, 0666)
			Expect(err).ToNot(HaveOccurred())
			err = parser.InterpolateAndParse(pathToManifest, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			rawAppManifest, executeErr = parser.RawAppManifest(appName)
		})

		AfterEach(func() {
			err := os.RemoveAll(pathToManifest)
			Expect(err).ToNot(HaveOccurred())
		})

		When("marshaling does not error", func() {

			It("returns just the app's manifest", func() {
				expectedExpandedPath, err := filepath.EvalSymlinks(tmpMyPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(string(rawAppManifest)).To(MatchYAML(fmt.Sprintf(`applications:
- name: spark
  memory: 1G
  instances: 2
  docker:
    username: experiment
  path: %s`, expectedExpandedPath)))
			})
		})

		When("The app is not present", func() {
			BeforeEach(func() {
				appName = "not-here"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(AppNotInManifestError{Name: "not-here"}))
				Expect(rawAppManifest).To(BeNil())
			})
		})

	})
})
