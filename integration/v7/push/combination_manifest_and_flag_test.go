package push

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with a simple manifest and flags", func() {
	DescribeTable("manifest conflicts with push flags",
		func(manifest map[string]interface{}, expectedOutput string, flags ...string) {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifest)

				pushCommandAndFlags := append([]string{PushCommandName}, flags...)
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, pushCommandAndFlags...)
				Eventually(session.Err).Should(Say(regexp.QuoteMeta(expectedOutput)))
				Eventually(session).Should(Exit(1))
			})
		},

		Entry("Manifest buildpacks with docker flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":       "app-name",
						"buildpacks": []string{"ruby_buildpack"},
					},
				},
			},
			"Incorrect Usage: The flag option --docker-image, -o cannot be used with the manifest property buildpacks",
			"--docker-image", "docker",
		),

		Entry("Manifest path with docker flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": "app-name",
						"path": "~",
					},
				},
			},
			"Incorrect Usage: The flag option --docker-image, -o cannot be used with the manifest property path",
			"--docker-image", "docker",
		),

		Entry("Manifest docker with droplet flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":   "app-name",
						"docker": map[string]interface{}{"image": "docker"},
					},
				},
			},
			"Incorrect Usage: The flag option --droplet cannot be used with the manifest property docker",
			"--droplet", os.TempDir(),
		),

		Entry("Manifest buildpacks with droplet flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":       "app-name",
						"buildpacks": []string{"ruby_buildpack"},
					},
				},
			},
			"Incorrect Usage: The flag option --droplet cannot be used with the manifest property buildpacks",
			"--droplet", os.TempDir(),
		),

		Entry("Manifest path with droplet flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name": "app-name",
						"path": os.TempDir(),
					},
				},
			},
			"Incorrect Usage: The flag option --droplet cannot be used with the manifest property path",
			"--droplet", os.TempDir(),
		),

		Entry("Manifest docker with buildpacks flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":   "app-name",
						"docker": map[string]interface{}{"image": "docker"},
					},
				},
			},
			"Incorrect Usage: The flag option --buildpack, -b cannot be used with the manifest property docker",
			"-b", "ruby_buildpack",
		),

		Entry("Manifest docker with path flag",
			map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":   "app-name",
						"docker": map[string]interface{}{"image": "docker"},
					},
				},
			},
			"Incorrect Usage: The flag option --path, -p cannot be used with the manifest property docker",
			"-p", os.TempDir(),
		),
	)

	When("pushing multiple apps from the manifest", func() {
		Context("manifest contains multiple apps and '--no-start' is provided", func() {
			var appName1, appName2 string

			BeforeEach(func() {
				Skip("After #162558994 has been completed")
				appName1 = helpers.NewAppName()
				appName2 = helpers.NewAppName()
			})

			It("does not start the apps", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]string{
							{"name": appName1},
							{"name": appName2},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say(`Applying manifest\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName1))
					Eventually(session).Should(Say(`requested state:\s+stopped`))
					Eventually(session).Should(Say(`name:\s+%s`, appName2))
					Eventually(session).Should(Say(`requested state:\s+stopped`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("manifest contains multiple apps and a '-p' is provided", func() {
			var tempDir string

			BeforeEach(func() {
				var err error
				tempDir, err = ioutil.TempDir("", "combination-manifest-with-p")
				Expect(err).ToNot(HaveOccurred())

				helpers.WriteManifest(filepath.Join(tempDir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]string{
						{
							"name": "name-1",
						},
						{
							"name": "name-2",
						},
					},
				})
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, PushCommandName, "-p", dir)
					Eventually(session.Err).Should(Say(regexp.QuoteMeta("Incorrect Usage: Command line flags (except -f and --no-start) cannot be applied when pushing multiple apps from a manifest file.")))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		DescribeTable("errors when any flag (except for -f and --no-start) is specified",
			func(flags ...string) {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]string{
							{"name": "some-app"},
							{"name": "some-other-app"},
						},
					})

					args := append([]string{PushCommandName}, flags...)
					session := helpers.CustomCF(helpers.CFEnv{
						WorkingDirectory: dir,
						EnvVars:          map[string]string{"CF_DOCKER_PASSWORD": "some-password"},
					}, args...)
					Eventually(session.Err).Should(Say(regexp.QuoteMeta("Incorrect Usage: Command line flags (except -f and --no-start) cannot be applied when pushing multiple apps from a manifest file.")))
					Eventually(session).Should(Exit(1))
				})
			},
			Entry("buildpack", "-b", "somethin"),
			Entry("disk", "-k", "100M"),
			Entry("docker image", "-o", "something"),
			Entry("docker image and username", "-o", "something", "--docker-username", "something"),
			Entry("health check timeout", "-t", "10"),
			Entry("health check type", "-u", "port"),
			Entry("instances", "-i", "10"),
			Entry("memory", "-m", "100M"),
			Entry("no route", "--no-route"),
			Entry("stack", "-s", "something"),
		)
	})

})
