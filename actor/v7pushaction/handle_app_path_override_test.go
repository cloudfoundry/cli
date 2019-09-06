package v7pushaction_test

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/pushmanifestparser"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HandleAppPathOverride", func() {
	var (
		transformedManifest pushmanifestparser.Manifest
		executeErr          error

		parsedManifest pushmanifestparser.Manifest
		flagOverrides  FlagOverrides
		err            error
	)

	BeforeEach(func() {
		flagOverrides = FlagOverrides{}
		parsedManifest = pushmanifestparser.Manifest{}
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		transformedManifest, executeErr = HandleAppPathOverride(
			parsedManifest,
			flagOverrides,
		)
	})

	When("the path flag override is set", func() {
		var filePath string

		BeforeEach(func() {
			file, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			filePath = file.Name()
			defer file.Close()

			flagOverrides = FlagOverrides{
				ProvidedAppPath: filePath,
			}
		})

		AfterEach(func() {
			err := os.RemoveAll(filePath)
			Expect(err).NotTo(HaveOccurred())
		})

		When("there are multiple apps in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{},
						{},
					},
				}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.CommandLineArgsWithMultipleAppsError{}))
			})
		})

		When("there is a single app in the manifest", func() {
			BeforeEach(func() {
				parsedManifest = pushmanifestparser.Manifest{
					Applications: []pushmanifestparser.Application{
						{
							Path: "some-non-existent-path",
						},
					},
				}
			})

			It("overrides the path for the first app in the manifest", func() {
				Expect(transformedManifest.Applications[0].Path).To(matchers.MatchPath(filePath))
			})
		})
	})

	When("the path flag override is not set", func() {
		BeforeEach(func() {
			parsedManifest = pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{
					{},
				},
			}
		})

		It("does not change the app path", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(transformedManifest.Applications[0].Path).To(Equal(""))
		})
	})

	When("the manifest contains an invalid path", func() {
		BeforeEach(func() {
			parsedManifest = pushmanifestparser.Manifest{
				Applications: []pushmanifestparser.Application{
					{
						Path: "some-non-existent-path",
					},
				},
			}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(pushmanifestparser.InvalidManifestApplicationPathError{
				Path: "some-non-existent-path",
			}))
		})
	})

	When("docker info is set in the manifest", func() {
		BeforeEach(func() {
			flagOverrides.ProvidedAppPath = "/some/path"

			parsedManifest.Applications = []pushmanifestparser.Application{
				{
					Name:   "some-app",
					Docker: &pushmanifestparser.Docker{},
				},
			}
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(translatableerror.ArgumentManifestMismatchError{
				Arg:              "--path, -p",
				ManifestProperty: "docker",
			}))
		})
	})
})
