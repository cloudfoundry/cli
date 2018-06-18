package pushaction_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifest"

	"github.com/cloudfoundry/bosh-cli/director/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReadManifest", func() {
	var (
		actor           *Actor
		fakeV2Actor     *pushactionfakes.FakeV2Actor
		fakeSharedActor *pushactionfakes.FakeSharedActor

		tmpDir         string
		manifestPath   string
		varsFilesPaths []string
		varsKV         []template.VarKV

		apps       []manifest.Application
		warnings   Warnings
		executeErr error
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeSharedActor = new(pushactionfakes.FakeSharedActor)
		actor = NewActor(fakeV2Actor, nil, fakeSharedActor)

		var err error
		tmpDir, err = ioutil.TempDir("", "read-manifest-test")
		Expect(err).ToNot(HaveOccurred())
		manifestPath = filepath.Join(tmpDir, "manifest.yml")

		varsFilesPaths = nil
		varsKV = nil
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	JustBeforeEach(func() {
		apps, warnings, executeErr = actor.ReadManifest(manifestPath, varsFilesPaths, varsKV)
	})

	Context("when provided `buildpack`", func() {
		BeforeEach(func() {
			manifest := []byte(`
---
applications:
  - name: some-app
    buildpack: some-buildpack
  - name: some-other-app
    buildpack: some-other-buildpack
`)

			Expect(ioutil.WriteFile(manifestPath, manifest, 0666)).To(Succeed())
		})

		It("sets the buildpack on the app and returns a deprecated field warning", func() {
			Expect(executeErr).ToNot(HaveOccurred())

			Expect(apps).To(ConsistOf(manifest.Application{
				Name:      "some-app",
				Buildpack: types.FilteredString{Value: "some-buildpack", IsSet: true},
			}, manifest.Application{
				Name:      "some-other-app",
				Buildpack: types.FilteredString{Value: "some-other-buildpack", IsSet: true},
			}))

			Expect(warnings).To(ConsistOf(`Deprecation warning: Use of 'buildpack' attribute in manifest is deprecated in favor of 'buildpacks'. Please see http://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#deprecated for alternatives and other app manifest deprecations. This feature will be removed in the future.`))
		})
	})
})
