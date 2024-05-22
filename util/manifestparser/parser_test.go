package manifestparser_test

import (
	"errors"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/manifestparser"
	"github.com/cloudfoundry/bosh-cli/director/template"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("ManifestParser", func() {
	var parser ManifestParser

	Describe("InterpolateManifest", func() {
		var (
			givenManifest    []byte
			pathToManifest   string
			pathsToVarsFiles []string
			vars             []template.VarKV

			interpolatedManifest []byte
			executeErr           error
		)

		BeforeEach(func() {
			tempFile, err := os.CreateTemp("", "manifest-test-")
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
			interpolatedManifest, executeErr = parser.InterpolateManifest(pathToManifest, pathsToVarsFiles, vars)
		})

		When("the manifest does *not* need interpolation", func() {
			BeforeEach(func() {
				givenManifest = []byte(`---
applications:
- name: spark
- name: flame
`)
				err := os.WriteFile(pathToManifest, givenManifest, 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("parses the manifest properly", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(string(interpolatedManifest)).To(Equal(`applications:
- name: spark
- name: flame
`))
			})
		})

		When("the manifest contains variables that need interpolation", func() {
			BeforeEach(func() {
				givenManifest = []byte(`---
applications:
- name: ((var1))
- name: ((var2))
`)
				err := os.WriteFile(pathToManifest, givenManifest, 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			When("only vars files are provided", func() {
				var (
					varsDir string
				)

				BeforeEach(func() {
					var err error
					varsDir, err = os.MkdirTemp("", "vars-test")
					Expect(err).ToNot(HaveOccurred())

					varsFilePath1 := filepath.Join(varsDir, "vars-1")
					err = os.WriteFile(varsFilePath1, []byte("var1: spark"), 0666)
					Expect(err).ToNot(HaveOccurred())

					varsFilePath2 := filepath.Join(varsDir, "vars-2")
					err = os.WriteFile(varsFilePath2, []byte("var2: flame"), 0666)
					Expect(err).ToNot(HaveOccurred())

					pathsToVarsFiles = append(pathsToVarsFiles, varsFilePath1, varsFilePath2)
				})

				AfterEach(func() {
					Expect(os.RemoveAll(varsDir)).ToNot(HaveOccurred())
				})

				When("multiple values for the same variable(s) are provided", func() {
					BeforeEach(func() {
						varsFilePath1 := filepath.Join(varsDir, "vars-1")
						err := os.WriteFile(varsFilePath1, []byte("var1: garbageapp\nvar1: spark\nvar2: doesn't matter"), 0666)
						Expect(err).ToNot(HaveOccurred())

						varsFilePath2 := filepath.Join(varsDir, "vars-2")
						err = os.WriteFile(varsFilePath2, []byte("var2: flame"), 0666)
						Expect(err).ToNot(HaveOccurred())

						pathsToVarsFiles = append(pathsToVarsFiles, varsFilePath1, varsFilePath2)
					})

					It("interpolates the placeholder values", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(string(interpolatedManifest)).To(Equal(`applications:
- name: spark
- name: flame
`))
					})
				})

				When("the provided files exists and contain valid yaml", func() {
					It("interpolates the placeholder values", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(string(interpolatedManifest)).To(Equal(`applications:
- name: spark
- name: flame
`))
					})
				})

				When("a variable in the manifest is not provided in the vars file", func() {
					BeforeEach(func() {
						varsFilePath := filepath.Join(varsDir, "vars-1")
						err := os.WriteFile(varsFilePath, []byte("notvar: foo"), 0666)
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
						err := os.WriteFile(varsFilePath, []byte(": bad"), 0666)
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
					Expect(string(interpolatedManifest)).To(Equal(`applications:
- name: spark
- name: flame
`))
				})
			})

			When("vars and vars files are provided", func() {
				var varsFilePath string
				BeforeEach(func() {
					tmp, err := os.CreateTemp("", "util-manifest-varsilfe")
					Expect(err).NotTo(HaveOccurred())
					Expect(tmp.Close()).NotTo(HaveOccurred())

					varsFilePath = tmp.Name()
					err = os.WriteFile(varsFilePath, []byte("var1: spark\nvar2: 12345"), 0666)
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
					Expect(string(interpolatedManifest)).To(Equal(`applications:
- name: spark
- name: flame
`))
				})
			})
		})
	})

	Describe("ParseManifest", func() {
		var (
			pathToManifest string
			rawManifest    []byte

			executeErr     error
			parsedManifest Manifest
		)

		BeforeEach(func() {
			pathToManifest = "/some/path/to/manifest.yml"
			rawManifest = nil
		})

		JustBeforeEach(func() {
			parsedManifest, executeErr = parser.ParseManifest(pathToManifest, rawManifest)
		})

		When("the manifest does not contain applications", func() {
			BeforeEach(func() {
				rawManifest = []byte(`applications:
`)
			})

			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(errors.New("Manifest must have at least one application.")))
			})
		})

		When("invalid yaml is passed", func() {
			BeforeEach(func() {
				rawManifest = []byte("\t\t")
			})

			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
			})
		})

		When("unmarshalling returns an error", func() {
			BeforeEach(func() {
				rawManifest = []byte(`---
	blah blah
	`)
			})

			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(&yaml.TypeError{}))
			})
		})

		When("the manifest is valid", func() {
			BeforeEach(func() {
				rawManifest = []byte(`applications:
- name: one
- name: two
`)
			})

			It("interpolates the placeholder values", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(parsedManifest.AppNames()).To(ConsistOf("one", "two"))
			})
		})
	})

	Describe("MarshalManifest", func() {
		It("marshals the manifest", func() {
			manifest := Manifest{
				Applications: []Application{
					{
						Name: "app-1",
						Processes: []Process{
							{
								Type:                    "web",
								RemainingManifestFields: map[string]interface{}{"unknown-process-key": 2},
							},
						},
						RemainingManifestFields: map[string]interface{}{"unknown-key": 1},
					},
				},
			}

			yaml, err := parser.MarshalManifest(manifest)

			Expect(err).NotTo(HaveOccurred())
			Expect(yaml).To(MatchYAML(`applications:
- name: app-1
  unknown-key: 1
  processes:
  - type: web
    unknown-process-key: 2
`))
		})
	})
})
