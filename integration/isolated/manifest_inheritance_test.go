package isolated

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// pushes app with multiple manifests, children being passed in first in the
// array
func pushHelloWorldAppWithManifests(manifests []string) {
	helpers.WithHelloWorldApp(func(appDir string) {
		pushPath := filepath.Join(appDir, "manifest-0.yml")
		for i, manifest := range manifests {
			manifestPath := filepath.Join(appDir, fmt.Sprintf("manifest-%d.yml", i))
			manifest = strings.Replace(manifest, "inherit: {some-parent}", fmt.Sprintf("inherit: manifest-%d.yml", i+1), 1)
			manifest = strings.Replace(manifest, "path: {some-dir}", fmt.Sprintf("path: %s", appDir), -1)
			err := ioutil.WriteFile(manifestPath, []byte(manifest), 0666)
			Expect(err).ToNot(HaveOccurred())
		}
		Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", "-f", pushPath)).Should(Exit(0))
	})
}

func pushDockerWithManifests(manifests []string) {
	helpers.WithHelloWorldApp(func(appDir string) {
		pushPath := filepath.Join(appDir, "manifest-0.yml")
		for i, manifest := range manifests {
			manifestPath := filepath.Join(appDir, fmt.Sprintf("manifest-%d.yml", i))
			manifest = strings.Replace(manifest, "inherit: {some-parent}", fmt.Sprintf("inherit: manifest-%d.yml", i+1), 1)
			manifest = strings.Replace(manifest, "path: {some-dir}", fmt.Sprintf("path: %s", appDir), -1)
			err := ioutil.WriteFile(manifestPath, []byte(manifest), 0666)
			Expect(err).ToNot(HaveOccurred())
		}
		Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", "-f", pushPath)).Should(Exit())
	})
}

// CASE 1: child manifest (no inheritance)
//
// APPLICATION params:
// values (memory, disk), list (routes), map (env vars)
//
// GLOBAL params:
// value, list, map types
//
// APPLICATION and GLOBAL params:
// value: application values override global values
// list: application lists append to global lists
// map: application maps merge & override global maps
//
//
// CASE 2: child + parent manifests (1 level inheritance)
//
// Parent Application & Child Application
// Parent Global & Child Application
// Parent Application & Child Global
// Parent Global & Child Global
//
// Parent Global & Child Global & Child Application
// Parent Application & Child Global & Child Application
// Parent Global & Parent Application & Child Application
// Parent Global & Parent Application & Child Global
//
// Parent Global & Parent Application & Child Global & Child Application
//
// CASE 3: child + parent + super-parent manifests (n+ inheritance)
// Super-parent Global & Parent Global & Parent Application & Child Global
// Super-parent Global & Parent Global & Child Global & Child Application

var _ = Describe("manifest inheritance in push command", func() {
	var (
		orgName     string
		spaceName   string
		domainName  string
		app1Name    string
		app2Name    string
		app1MemSize int
		app2MemSize int
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		app1Name = helpers.PrefixedRandomName("app")
		app2Name = helpers.PrefixedRandomName("app")
		app1MemSize = 32
		app2MemSize = 32

		setupCF(orgName, spaceName)

		domainName = fmt.Sprintf("%s.com", helpers.PrefixedRandomName("DOMAIN"))
		helpers.NewDomain(orgName, domainName).Create()
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	Context("when there is only one manifest", func() {
		Context("when the manifest contains only applications properties", func() {
			BeforeEach(func() {
				pushHelloWorldAppWithManifests([]string{fmt.Sprintf(`
---
applications:
- name: %s
  memory: %dM
  disk_quota: 128M
  buildpack: staticfile_buildpack
  path: {some-dir}
  routes:
  - route: hello.%s
  - route: hi.%s
  env:
    BAR: bar
    FOO: foo
- name: %s
  memory: %dM
  disk_quota: 128M
  buildpack: staticfile_buildpack
  path: {some-dir}
  routes:
  - route: hello.%s
  - route: hi.%s
  env:
    BAR: bar
    FOO: foo
`, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName)})
			})

			It("pushes the same applications properties", func() {
				session := helpers.CF("env", app1Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "hello\.%s"\,
   "hi\.%s"
  \]`, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 128`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: bar
FOO: foo

`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("env", app2Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "hello\.%s"\,
   "hi\.%s"
  \]`, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 128`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: bar
FOO: foo

`))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the manifest contains mainly global properties", func() {
			BeforeEach(func() {
				pushHelloWorldAppWithManifests([]string{fmt.Sprintf(`
---
memory: %dM
disk_quota: 128M
buildpack: staticfile_buildpack
path: {some-dir}
routes:
- route: hello.%s
- route: hi.%s
env:
  BAR: bar
  FOO: foo
applications:
- name: %s
- name: %s
`, app1MemSize, domainName, domainName, app1Name, app2Name)})
			})

			It("pushes the same global properties", func() {
				session := helpers.CF("env", app1Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "hello\.%s"\,
   "hi\.%s"
  \]`, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 128`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: bar
FOO: foo

`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("env", app2Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "hello\.%s"\,
   "hi\.%s"
  \]`, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 128`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: bar
FOO: foo

`))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the manifest contains both applications and global properties", func() {
			BeforeEach(func() {
				pushHelloWorldAppWithManifests([]string{fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 64M
disk_quota: 256M
routes:
- route: global-1.%s
- route: global-2.%s
env:
  BAR: global
  FOO: global
applications:
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: app-1.%s
  - route: app-2.%s
  env:
    BAR: app
    BAZ: app
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: app-1.%s
  - route: app-2.%s
  env:
    BAR: app
    BAZ: app
`, domainName, domainName, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName)})
			})

			It("pushes with application properties taking precedence; values are overwritten, lists are appended, and maps are merged", func() {
				session := helpers.CF("env", app1Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "global-1\.%s"\,
   "global-2\.%s",
   "app-1\.%s",
   "app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 128`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: app
BAZ: app
FOO: global

`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("env", app2Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "global-1\.%s"\,
   "global-2\.%s",
   "app-1\.%s",
   "app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 128`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: app
BAZ: app
FOO: global

`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when there are two manifests", func() {
		Context("when the child has applications properties; and the parent has applications properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    BAZ: child-app
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    BAZ: child-app
`, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
						fmt.Sprintf(`
---
applications:
- name: %s
  buildpack: staticfile_buildpack
  memory: %dM
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    BAZ: parent-app
    FOO: parent-app
- name: %s
  buildpack: staticfile_buildpack
  memory: %dM
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    BAZ: parent-app
    FOO: parent-app
`, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
					})
				})

				It("pushes with child application properties taking precedence; values are overwritten, lists are appended, and maps are merged", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-app-1\.%s",
   "parent-app-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: child-app
FOO: parent-app

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-app-1\.%s",
   "parent-app-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: child-app
FOO: parent-app

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  docker:
    image: child-application-image
`, app1Name),
						fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: %s
`, app1Name, DockerImage),
					})
				})

				It("pushes with child application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has applications properties; and the parent has global properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    BAZ: child-app
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    BAZ: child-app
`, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 64M
disk_quota: 256M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  BAR: parent-global
  BAZ: parent-global
  FOO: parent-global
`, domainName, domainName),
					})
					SetDefaultEventuallyTimeout(300 * time.Second)
				})

				It("pushes with child application properties taking precedence", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: child-app
FOO: parent-global

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: child-app
FOO: parent-global

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  docker:
    image: child-application-image
`, app1Name),
						fmt.Sprintf(`
---
docker:
  image: parent-global-image
`),
					})
				})

				It("pushes with child application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has global properties; and the parent has applications properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
buildpack: staticfile_buildpack
memory: 64M
disk_quota: 256M
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  BAR: child-global
  BAZ: child-global
`, domainName, domainName),
						fmt.Sprintf(`
---
applications:
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    BAZ: parent-app
    FOO: parent-app
- name: %s
  memory: %dM
  disk_quota: 128M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    BAZ: parent-app
    FOO: parent-app
`, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
					})
					SetDefaultEventuallyTimeout(300 * time.Second)
				})

				It("pushes with parent application properties taking precedence", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "child-global-1\.%s"\,
   "child-global-2\.%s",
   "parent-app-1\.%s",
   "parent-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: parent-app
BAZ: parent-app
FOO: parent-app
`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "child-global-1\.%s"\,
   "child-global-2\.%s",
   "parent-app-1\.%s",
   "parent-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: parent-app
BAZ: parent-app
FOO: parent-app
`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
docker:
  image: child-global-image
`),
						fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: parent-application-image
`, app1Name),
					})
				})

				It("pushes with parent application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*parent-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has global properties; and the parent has global properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
memory: %dM
disk_quota: 128M
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  BAR: child-global
  FOO: child-global
`, app1MemSize, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 64M
disk_quota: 256M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  BAR: parent-global
  FOO: parent-global
  BAZ: parent-global
applications:
- name: %s
- name: %s
`, domainName, domainName, app1Name, app2Name),
					})
				})

				It("pushes with child global properties taking precedence;", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s"\,
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-global
BAZ: parent-global
FOO: child-global

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s"\,
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s"
  \]`, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 128`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-global
BAZ: parent-global
FOO: child-global

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
docker:
  image: child-global-image
`),
						fmt.Sprintf(`
---
docker:
  image: %s
applications:
- name: %s
`, DockerImage, app1Name),
					})
				})

				It("pushes with child global docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-global-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has applications and global properties; and the parent has global properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
memory: 128M
disk_quota: 128M
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  FOO: child-global
  FIZ: child-global
applications:
- name: %s
  memory: %dM
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    FOO: child-app
- name: %s
  memory: %dM
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    FOO: child-app
`, domainName, domainName, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 256M
disk_quota: 256M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  BAR: parent-global
  FOO: parent-global
  FIZ: parent-global
  BAZ: parent-global
applications:
- name: %s
- name: %s
`, domainName, domainName, app1Name, app2Name),
					})
				})

				It("pushes with child application taking precedence over child global over parent global", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s"\,
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "child-app-1\.%s",
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-global
FIZ: child-global
FOO: child-app

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s"\,
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "child-app-1\.%s",
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-global
FIZ: child-global
FOO: child-app

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  docker:
    image: child-application-image
docker:
  image: child-global-image
`, app1Name),
						fmt.Sprintf(`
---
docker:
  image: parent-image
`),
					})
				})

				It("pushes with child application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has applications and global properties; and the parent has applications properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
memory: %dM
disk_quota: 128M
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  FOO: child-global
  FIZ: child-global
applications:
- name: %s
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    FOO: child-app
- name: %s
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    FOO: child-app
`, app1MemSize, domainName, domainName, app1Name, domainName, domainName, app2Name, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
applications:
- name: %s
  memory: 256M
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    FOO: parent-app
    FIZ: parent-app
    BAZ: parent-app
- name: %s
  memory: 256M
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    FOO: parent-app
    FIZ: parent-app
    BAZ: parent-app
`, app1Name, domainName, domainName, app2Name, domainName, domainName),
					})
				})

				It("pushes with child application taking precedence over child global over parent application", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s",
   "child-app-1\.%s",
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-app
FIZ: child-global
FOO: child-app

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s",
   "child-app-1\.%s",
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-app
FIZ: child-global
FOO: child-app

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  docker:
    image: child-application-image
docker:
  image: child-global-image
`, app1Name),
						fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: %s
`, app1Name, DockerImage),
					})
				})

				It("pushes with child application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has applications properties; and the parent has applications and global properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  memory: %dM
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    FOO: child-app
- name: %s
  memory: %dM
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    BAR: child-app
    FOO: child-app
`, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 128M
disk_quota: 128M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  FOO: parent-global
  FIZ: parent-global
applications:
- name: %s
  memory: 256M
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    FOO: parent-app
    FIZ: parent-app
    BAZ: parent-app
- name: %s
  memory: 256M
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    BAR: parent-app
    FOO: parent-app
    FIZ: parent-app
    BAZ: parent-app
`, domainName, domainName, app1Name, domainName, domainName, app2Name, domainName, domainName),
					})
				})

				It("pushes with child application taking precedence over parent application over child global", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s",
   "child-app-1\.%s",
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-app
FIZ: parent-global
FOO: child-app

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s",
   "child-app-1\.%s",
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-app
FIZ: parent-global
FOO: child-app

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  docker:
    image: child-application-image
`, app1Name),
						fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: %s
docker:
  image: %s
`, app1Name, DockerImage, DockerImage),
					})
				})

				It("pushes with child application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has global properties; and the parent has applications and global properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
memory: 128M
disk_quota: 128M
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  FOO: child-global
  FIZ: child-global
`, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  FIZ: parent-global
  ZOOM: parent-global
applications:
- name: %s
  memory: %dM
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    FOO: parent-app
    BAZ: parent-app
- name: %s
  memory: %dM
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    FOO: parent-app
    BAZ: parent-app
`, domainName, domainName, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
					})
				})

				It("pushes with parent application taking precedence over child global over parent global", func() {
					session := helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 256`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAZ: parent-app
FIZ: child-global
FOO: parent-app
ZOOM: parent-global

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 256`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAZ: parent-app
FIZ: child-global
FOO: parent-app
ZOOM: parent-global

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
docker:
  image: child-global-image
`),
						fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: parent-application-image
docker:
  image: parent-global-image
`, app1Name),
					})
				})

				It("pushes with parent application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*parent-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the child has applications and global properties; and the parent has applications and global properties", func() {
			Context("when the app is a buildpack app", func() {
				BeforeEach(func() {
					pushHelloWorldAppWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
instances: 2
memory: 128M
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  FOO: child-global
  FIZ: child-global
applications:
- name: %s
  memory: %dM
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    FOO: child-app
    BAR: child-app
- name: %s
  memory: %dM
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    FOO: child-app
    BAR: child-app
`, domainName, domainName, app1Name, app1MemSize, domainName, domainName, app2Name, app2MemSize, domainName, domainName),
						fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 256M
disk_quota: 256M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  FIZ: parent-global
  ZOOM: parent-global
applications:
- name: %s
  memory: 256M
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    FIZ: parent-app
    BAZ: parent-app
- name: %s
  memory: 256M
  disk_quota: 256M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    FIZ: parent-app
    BAZ: parent-app
`, domainName, domainName, app1Name, domainName, domainName, app2Name, domainName, domainName),
					})
				})

				It("pushes with parent application taking precedence over child global", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say("instances.*2"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app1Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-app
FIZ: child-global
FOO: child-app
ZOOM: parent-global

`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("app", app2Name)
					Eventually(session.Out).Should(Say("instances.*2"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("env", app2Name)
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say(`"application_uris": \[
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName, domainName, domainName))
					Eventually(session.Out).Should(Say(`"disk": 64`))
					Eventually(session.Out).Should(Say(`"mem": %d`, app2MemSize))
					Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
BAZ: parent-app
FIZ: child-global
FOO: child-app
ZOOM: parent-global

`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is a Docker app", func() {
				BeforeEach(func() {
					pushDockerWithManifests([]string{
						fmt.Sprintf(`
---
inherit: {some-parent}
applications:
- name: %s
  docker:
    image: child-application-image
docker:
  image: child-global-image
`, app1Name),
						fmt.Sprintf(`
---
applications:
- name: %s
  docker:
    image: %s
docker:
  image: %s
`, app1Name, DockerImage, DockerImage),
					})
				})

				It("pushes with child application docker image properties taking precedence", func() {
					session := helpers.CF("app", app1Name)
					Eventually(session.Out).Should(Say(`docker image:\s*child-application-image`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	Context("when there are 3 manifests", func() {
		Context("when super-parent has globals, parent has globals, and child has app and globals", func() {
			BeforeEach(func() {
				pushHelloWorldAppWithManifests([]string{
					fmt.Sprintf(`
---
inherit: {some-parent}
memory: %dM
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  FOO: child-global
  FIZ: child-global
applications:
- name: %s
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    FOO: child-app
    BAR: child-app
- name: %s
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: child-app-1.%s
  - route: child-app-2.%s
  env:
    FOO: child-app
    BAR: child-app
`, app1MemSize, domainName, domainName, app1Name, domainName, domainName, app2Name, domainName, domainName),
					fmt.Sprintf(`
---
inherit: {some-parent}
instances: 2
memory: 256M
disk_quota: 256M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  ZOOM: parent-global
  FIZ: parent-global
`, domainName, domainName),
					fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 512M
disk_quota: 512M
path: {some-dir}
routes:
- route: super-parent-global-1.%s
- route: super-parent-global-2.%s
env:
  MOON: super-parent-global
  ZOOM: super-parent-global
`, domainName, domainName),
				})
			})

			It("pushes with child application taking precedence over child global over parent global over super-parent global", func() {
				session := helpers.CF("app", app1Name)
				Eventually(session.Out).Should(Say("instances.*2"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("env", app1Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "super-parent-global-1\.%s",
   "super-parent-global-2\.%s",
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 64`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
FIZ: child-global
FOO: child-app
MOON: super-parent-global
ZOOM: parent-global

`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", app2Name)
				Eventually(session.Out).Should(Say("instances.*2"))

				session = helpers.CF("env", app2Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "super-parent-global-1\.%s",
   "super-parent-global-2\.%s",
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "child-app-1\.%s"\,
   "child-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 64`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: child-app
FIZ: child-global
FOO: child-app
MOON: super-parent-global
ZOOM: parent-global

`))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when super-parent has globals, parent has app and globals, and child has globals", func() {
			BeforeEach(func() {
				pushHelloWorldAppWithManifests([]string{
					fmt.Sprintf(`
---
inherit: {some-parent}
memory: %dM
path: {some-dir}
routes:
- route: child-global-1.%s
- route: child-global-2.%s
env:
  FOO: child-global
  FIZ: child-global
  JUNE: child-global
`, app1MemSize, domainName, domainName),
					fmt.Sprintf(`
---
inherit: {some-parent}
instances: 2
memory: 256M
disk_quota: 256M
path: {some-dir}
routes:
- route: parent-global-1.%s
- route: parent-global-2.%s
env:
  ZOOM: parent-global
  FIZ: parent-global
applications:
- name: %s
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    FOO: parent-app
    BAR: parent-app
- name: %s
  disk_quota: 64M
  path: {some-dir}
  routes:
  - route: parent-app-1.%s
  - route: parent-app-2.%s
  env:
    FOO: parent-app
    BAR: parent-app
`, domainName, domainName, app1Name, domainName, domainName, app2Name, domainName, domainName),
					fmt.Sprintf(`
---
buildpack: staticfile_buildpack
memory: 512M
disk_quota: 512M
path: {some-dir}
routes:
- route: super-parent-global-1.%s
- route: super-parent-global-2.%s
env:
  MOON: super-parent-global
  ZOOM: super-parent-global
  JUNE: super-parent-global
`, domainName, domainName),
				})
			})

			It("pushes with child application taking precedence over child global over parent global over super-parent global", func() {
				session := helpers.CF("app", app1Name)
				Eventually(session.Out).Should(Say("instances.*2"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("env", app1Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "super-parent-global-1\.%s",
   "super-parent-global-2\.%s",
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 64`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: parent-app
FIZ: child-global
FOO: parent-app
JUNE: child-global
MOON: super-parent-global
ZOOM: parent-global

`))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", app1Name)
				Eventually(session.Out).Should(Say("instances.*2"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("env", app2Name)
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(`"application_uris": \[
   "super-parent-global-1\.%s",
   "super-parent-global-2\.%s",
   "parent-global-1\.%s",
   "parent-global-2\.%s",
   "child-global-1\.%s",
   "child-global-2\.%s",
   "parent-app-1\.%s"\,
   "parent-app-2\.%s"
  \]`, domainName, domainName, domainName, domainName, domainName, domainName, domainName, domainName))
				Eventually(session.Out).Should(Say(`"disk": 64`))
				Eventually(session.Out).Should(Say(`"mem": %d`, app1MemSize))
				Eventually(session.Out).Should(Say(`User-Provided:
BAR: parent-app
FIZ: child-global
FOO: parent-app
JUNE: child-global
MOON: super-parent-global
ZOOM: parent-global

`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
