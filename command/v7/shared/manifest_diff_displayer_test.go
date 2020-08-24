package shared_test

import (
	"fmt"

	. "code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("ManifestDiffDisplayer", func() {
	var (
		testUI       *ui.UI
		outputBuffer *Buffer
		displayer    *ManifestDiffDisplayer
	)

	BeforeEach(func() {
		outputBuffer = NewBuffer()
		testUI = ui.NewTestUI(nil, outputBuffer, NewBuffer())
		displayer = NewManifestDiffDisplayer(testUI)
	})

	Describe("DisplayDiff", func() {
		var (
			rawManifest []byte
			diff        resources.ManifestDiff
		)

		JustBeforeEach(func() {
			displayer.DisplayDiff(rawManifest, diff)
		})

		Context("No diffs", func() {
			BeforeEach(func() {
				rawManifest = []byte(`---
version: 1
applications:
- name: app1
  buildpacks:
  - ruby_buildpack
  - java_buildpack
  env:
    VAR1: value1
    VAR2: value2
  routes:
  - route: route.example.com
  - route: another-route.example.com
  services:
  - my-service1
  - my-service2
  - name: my-service-with-arbitrary-params
    parameters:
      key1: value1
      key2: value2
  stack: cflinuxfs3
  metadata:
    annotations:
      contact: "bob@example.com jane@example.com"
    labels:
      sensitive: true
  processes:
  - type: web
    command: start-web.sh
    disk_quota: 512M
    health-check-http-endpoint: /healthcheck
    health-check-type: http
    health-check-invocation-timeout: 10
    instances: 3
    memory: 500M
    timeout: 10
  - type: worker
    command: start-worker.sh
    disk_quota: 1G
    health-check-type: process
    instances: 2
    memory: 256M
    timeout: 15`)
			})

			It("outputs the manifest without + or -", func() {
				println(string(rawManifest))
				println("000000000000")
				println(string(outputBuffer.Contents()))
				Expect(string(outputBuffer.Contents())).To(MatchRegexp(`---
version: 1
applications:
- name: app1
  buildpacks:
  - ruby_buildpack
  - java_buildpack
  env:
    VAR1: value1
    VAR2: value2
  routes:
  - route: route.example.com
  - route: another-route.example.com
  services:
  - my-service1
  - my-service2
  - name: my-service-with-arbitrary-params
    parameters:
      key1: value1
      key2: value2
  stack: cflinuxfs3
  metadata:
    annotations:
      contact: "bob@example.com jane@example.com"
    labels:
      sensitive: true
  processes:
  - type: web
    command: start-web.sh
    disk_quota: 512M
    health-check-http-endpoint: /healthcheck
    health-check-type: http
    health-check-invocation-timeout: 10
    instances: 3
    memory: 500M
    timeout: 10
  - type: worker
    command: start-worker.sh
    disk_quota: 1G
    health-check-type: process
    instances: 2
    memory: 256M
    timeout: 15`))
			})
		})

		Context("Operation kinds", func() {
			When("add", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: b`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/0/env/r", Value: "m"},
						},
					}
				})

				FIt("outputs a diff indicating addition", func() {
					Expect(string(outputBuffer.Contents())).To(MatchRegexp(`---
  applications:
  -  name: dora
     env:
       a: b
  +    r: m`))
				})
			})

			When("remove", func() {
				BeforeEach(func() {
					rawManifest = []byte("applications:\n- name: dora\n  env:\n    r: m")
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.RemoveOperation, Path: "/applications/0/env/a", Was: "b"},
						},
					}
				})

				It("outputs correctly formatted diff", func() {
					// TODO: Remove printline, currently just printing test output
					fmt.Printf("%+v", string(outputBuffer.Contents()))

					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    name: dora
    env:
-     a: b
      r: m
`))
				})
			})

			When("replace", func() {
				BeforeEach(func() {
					rawManifest = []byte("applications:\n- name: dora\n  env:\n    a: c\n    r: m")
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.ReplaceOperation, Path: "/applications/0/env/a", Was: "b", Value: "c"},
						},
					}
				})

				It("outputs correctly formatted diff", func() {
					// TODO: Remove printline, currently just printing test output
					fmt.Printf("%+v", string(outputBuffer.Contents()))

					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    name: dora
    env:
-     a: b
+     a: c
      r: m
`))
				})
			})

		})

		Context("Edge Cases", func() {
			When("the diff value contains a map", func() {
				BeforeEach(func() {
					rawManifest = []byte(`applications:
    -
      name: dora
      env:
        a: b
        r: m
    -
      name: dora1
      env:
        new: variable
`)

					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/1/name", Value: "dora1"},
							{Op: resources.AddOperation, Path: "/applications/1/env", Value: map[string]interface{}{"new": "variable"}},
						},
					}
				})

				It("outputs correctly formatted diff for array", func() {
					// TODO: Remove printline, currently just printing test output
					fmt.Printf("%+v", string(outputBuffer.Contents()))

					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    name: dora
    env:
      a: b
      r: m
+   name: dora1
+   env:
+     new: variable
`))
				})
			})

			When("The diff value contains an array of maps", func() {
				BeforeEach(func() {
					rawManifest = []byte(`applications:
    -
      name: dora
      routes:
      - route: new-route.com
      - route: another-new-route.com
`)

					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/0/routes", Value: []map[string]interface{}{{"route": "new-route.com"}, {"route": "another-new-route.com"}}},
						},
					}
				})

				It("outputs correctly formatted diff for map", func() {
					// TODO: Remove printline, currently just printing test output
					fmt.Printf("%+v", string(outputBuffer.Contents()))

					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    name: dora
+   routes:
+     route: new-route.com
+     route: another-new-route.com
`))
				})
			})

			When("The diff value contains a multiline field", func() {
			})
		})
	})
})
