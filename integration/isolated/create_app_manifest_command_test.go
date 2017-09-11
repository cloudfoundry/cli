package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func createManifest(appName string) (manifest.Manifest, error) {
	tmpDir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return manifest.Manifest{}, err
	}

	manifestPath := filepath.Join(tmpDir, "manifest.yml")
	Eventually(helpers.CF("create-app-manifest", appName, "-p", manifestPath)).Should(Exit(0))

	manifestContents, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return manifest.Manifest{}, err
	}

	var appsManifest manifest.Manifest
	err = yaml.Unmarshal(manifestContents, appsManifest)
	if err != nil {
		return manifest.Manifest{}, err
	}

	return appsManifest, nil
}

var _ = Describe("create-app-manifest command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()

		appName = helpers.PrefixedRandomName("app")

		setupCF(orgName, spaceName)
	})

	Context("when app has no hostname", func() {
		var domain helpers.Domain

		BeforeEach(func() {
			domain = helpers.NewDomain(orgName, helpers.DomainName(""))
			domain.Create()

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-hostname", "-d", domain.Name)).Should(Exit(0))
			})
		})

		It("contains routes without hostnames", func() {
			manifest, err := createManifest(appName)
			Expect(err).ToNot(HaveOccurred())

			Expect(manifest.Applications).To(HaveLen(1))
			Expect(manifest.Applications[0].Routes).To(HaveLen(1))
			Expect(manifest.Applications[0].Routes[0]).To(Equal(domain.Name))
		})
	})

	Context("health check type", func() {
		Context("when the health check type is port", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "port")).Should(Exit(0))
				})
			})

			It("does not write the type or endpoint to the manifest", func() {
				manifest, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(manifest.Applications).To(HaveLen(1))
				Expect(manifest.Applications[0].HealthCheckType).To(BeEmpty())
				Expect(manifest.Applications[0].HealthCheckHTTPEndpoint).To(BeEmpty())
			})

			Context("when the health check http endpoint is not /", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
					Eventually(helpers.CF("set-health-check", appName, "port")).Should(Exit(0))
				})

				It("still does not write the endpoint to the manifest", func() {
					manifest, err := createManifest(appName)
					Expect(err).ToNot(HaveOccurred())

					Expect(manifest.Applications).To(HaveLen(1))
					Expect(manifest.Applications[0].HealthCheckHTTPEndpoint).To(BeEmpty())
				})
			})
		})

		Context("when the health check type is not port", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
				})
			})

			It("writes it to the manifest", func() {
				manifest, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(manifest.Applications).To(HaveLen(1))
				Expect(manifest.Applications[0].HealthCheckType).To(Equal("http"))
			})
		})
	})

	Context("health check http endpoint", func() {
		BeforeEach(func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
			})
		})

		Context("when the health check http endpoint is /", func() {
			It("does not write it to the manifest", func() {
				manifest, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(manifest.Applications).To(HaveLen(1))
				Expect(manifest.Applications[0].HealthCheckHTTPEndpoint).To(BeEmpty())
			})
		})

		Context("when the health check endpoint is not /", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
			})

			It("writes it to the manifest", func() {
				manifest, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(manifest.Applications).To(HaveLen(1))
				Expect(manifest.Applications[0].HealthCheckHTTPEndpoint).To(Equal("/some-endpoint"))
			})
		})
	})
})
