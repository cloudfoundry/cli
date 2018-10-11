// +build windows

package push

import (
	"path/filepath"
	"regexp"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing a path with the -p flag", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("pushing a relative root path (\\)", func() {
		It("pushes the app from the directory", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				volumeName := filepath.VolumeName(appDir)
				relativeRoot := strings.TrimPrefix(appDir, volumeName)
				Expect(strings.HasPrefix(relativeRoot, `\`))

				session := helpers.CF(PushCommandName, appName, "-p", relativeRoot)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("path:\\s+%s", regexp.QuoteMeta(appDir)))
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Comparing local files to remote cache\\.\\.\\."))
				Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))
				Eventually(session).Should(Say("name:\\s+%s", appName))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})

var _ = XDescribe("pushing a path from a manifest", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("pushing a relative root path (\\)", func() {
		It("pushes the app from the directory", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				volumeName := filepath.VolumeName(appDir)
				relativeRoot := strings.TrimPrefix(appDir, volumeName)
				Expect(strings.HasPrefix(relativeRoot, `\`))
				helpers.WriteManifest(filepath.Join(appDir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"path": relativeRoot,
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, PushCommandName)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("path:\\s+%s", regexp.QuoteMeta(appDir)))
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Comparing local files to remote cache\\.\\.\\."))
				Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))
				Eventually(session).Should(Say("name:\\s+%s", appName))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
