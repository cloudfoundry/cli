package shared_test

import (
	. "code.cloudfoundry.org/cli/v8/command/v7/shared"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("ManifestDiffDisplayer", func() {
	var (
		testUI    *ui.UI
		displayer *ManifestDiffDisplayer
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		displayer = &ManifestDiffDisplayer{
			UI:        testUI,
			RedactEnv: false,
		}
	})

	Describe("DisplayDiff", func() {
		var (
			rawManifest []byte
			diff        resources.ManifestDiff
			err         error
		)

		BeforeEach(func() {
			rawManifest = []byte("")
		})

		JustBeforeEach(func() {
			err = displayer.DisplayDiff(rawManifest, diff)
		})

		It("does not return an error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		Context("No diffs", func() {
			BeforeEach(func() {
				diff = resources.ManifestDiff{}
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
  stack: cflinuxfs4
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
				Expect(testUI.Out).To(Say(`---
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
    stack: cflinuxfs4
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
			When("adding a string value", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: b
    r: m`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/0/env/r", Value: "m"},
						},
					}
				})

				It("outputs a diff indicating addition for a single line", func() {
					Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    env:
      a: b
\+     r: m`))
				})
			})
			When("adding a map value within an array", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: b`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/0", Value: map[string]interface{}{
								"name": "dora",
								"env": map[string]interface{}{
									"a": "b",
								},
							}},
						},
					}
				})

				It("outputs a diff indicating addition of a map type", func() {
					Expect(testUI.Out).To(Say(`  ---
  applications:
\+ - env:
\+     a: b
\+   name: dora`))
				})
			})

			When("adding a map value within a map", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: b`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/0/env", Value: map[string]interface{}{
								"a": "b",
							}},
						},
					}
				})

				It("outputs a diff indicating addition of a map type", func() {
					Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
\+   env:
\+     a: b`))
				})
			})

			When("adding an array value", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    r: m
  routes:
  - route: route1.cli.fun
  - route: route2.cli.fun`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{
								Op:   resources.AddOperation,
								Path: "/applications/0/routes",
								Value: []map[string]interface{}{
									{
										"route": "route1.cli.fun",
									},
									{
										"route": "route2.cli.fun",
									},
								},
							},
						},
					}
				})

				When("each element of the array is a map value", func() {
					It("outputs a diff indicating addition for each map type", func() {
						Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    env:
      r: m
\+   routes:
\+   - route: route1.cli.fun
\+   - route: route2.cli.fun`))
					})
				})
			})

			When("remove", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    r: m`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.RemoveOperation, Path: "/applications/0/env/a", Was: "b"},
						},
					}
				})

				It("outputs correctly formatted diff with key removed", func() {
					Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    env:
      r: m
-     a: b`))
				})
			})

			When("replace", func() {
				BeforeEach(func() {
					rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: c
    r: m`)
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.ReplaceOperation, Path: "/applications/0/env/a", Was: "b", Value: "c"},
						},
					}
				})

				It("outputs correctly formatted diff", func() {
					Expect(testUI.Out).To(Say(`---
  applications:
  - name: dora
    env:
-     a: b
\+     a: c
      r: m
`))
				})
			})
		})

		Context("when the YAML cannot be parsed", func() {
			BeforeEach(func() {
				diff = resources.ManifestDiff{
					Diffs: []resources.Diff{
						{Op: resources.ReplaceOperation, Path: "/applications/0/env/a", Was: "b", Value: "c"},
					},
				}
				rawManifest = []byte(`not-real-yaml!`)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("unable to process manifest diff because its format is invalid"))
			})
		})
	})

	When("redacting is enabled", func() {
		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			displayer = &ManifestDiffDisplayer{
				UI:        testUI,
				RedactEnv: true,
			}
		})

		Describe("DisplayDiff", func() {
			var (
				rawManifest []byte
				diff        resources.ManifestDiff
				err         error
			)

			BeforeEach(func() {
				rawManifest = []byte("")
			})

			JustBeforeEach(func() {
				err = displayer.DisplayDiff(rawManifest, diff)
			})

			It("does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			Context("No diffs", func() {
				BeforeEach(func() {
					diff = resources.ManifestDiff{}
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
    TEST: |
      crazy things
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
  stack: cflinuxfs4
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
					Expect(testUI.Out).To(Say(`---
  version: 1
  applications:
  - name: app1
    buildpacks:
    - ruby_buildpack
    - java_buildpack
    env:
      VAR1: <redacted>
      VAR2: <redacted>
      TEST: <redacted>
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
    stack: cflinuxfs4
    metadata:
      annotations:
        contact: bob@example.com jane@example.com
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
      timeout: 15`,
					))
				})
			})

			Context("Operation kinds", func() {
				When("adding a string value", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: b
    r: m`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{Op: resources.AddOperation, Path: "/applications/0/env/r", Value: "m"},
							},
						}
					})

					It("outputs a diff indicating addition for a single line", func() {
						Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    env:
      a: <redacted>
\+     r: <redacted>`))
					})
				})
				When("adding a map value within an array", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{
									Op:    resources.AddOperation,
									Path:  "/applications/0",
									Was:   nil,
									Value: map[string]interface{}{"env": map[string]interface{}{"a": "b"}},
								},
							},
						}
					})

					It("outputs a diff indicating addition of a map type", func() {
						Expect(testUI.Out).To(Say(`  ---
  applications:
\+ - env:
\+     a: <redacted>
`))
					})
				})

				When("adding a map value within a map", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: b`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{Op: resources.AddOperation, Path: "/applications/0/env", Value: map[string]interface{}{
									"a": "b",
								}},
							},
						}
					})

					It("outputs a diff indicating addition of a map type", func() {
						Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
\+   env:
\+     a: <redacted>`))
					})
				})

				When("adding an array value", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
  env:
    r: m
  routes:
  - route: route1.cli.fun
  - route: route2.cli.fun`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{
									Op:   resources.AddOperation,
									Path: "/applications/0/routes",
									Value: []map[string]interface{}{
										{
											"route": "route1.cli.fun",
										},
										{
											"route": "route2.cli.fun",
										},
									},
								},
							},
						}
					})

					When("each element of the array is a map value", func() {
						It("outputs a diff indicating addition for each map type", func() {
							Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    env:
      r: <redacted>
\+   routes:
\+   - route: route1.cli.fun
\+   - route: route2.cli.fun`))
						})
					})
				})

				When("adding cnb-credentials", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
  lifecycle: cnb
  cnb-credentials:
   foo: bar`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{
									Op:   resources.AddOperation,
									Path: "/applications/0/cnb-credentials",
									Value: []map[string]interface{}{
										{
											"foo": "Bar",
										},
									},
								},
							},
						}
					})

					It("redacts output", func() {
						Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    lifecycle: cnb
\+   cnb-credentials: '\[PRIVATE DATA HIDDEN\]'`))
					})
				})

				When("remove", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
  env:
    r: m`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{Op: resources.RemoveOperation, Path: "/applications/0/env/a", Was: "b"},
							},
						}
					})

					It("outputs correctly formatted diff with key removed", func() {
						Expect(testUI.Out).To(Say(`  ---
  applications:
  - name: dora
    env:
      r: <redacted>
-     a: <redacted>`))
					})
				})

				When("replace", func() {
					BeforeEach(func() {
						rawManifest = []byte(`---
applications:
- name: dora
  env:
    a: c
    r: m`)
						diff = resources.ManifestDiff{
							Diffs: []resources.Diff{
								{Op: resources.ReplaceOperation, Path: "/applications/0/env/a", Was: "b", Value: "c"},
							},
						}
					})

					It("outputs correctly formatted diff", func() {
						Expect(testUI.Out).To(Say(`---
  applications:
  - name: dora
    env:
-     a: <redacted>
\+     a: <redacted>
      r: <redacted>
`))
					})
				})
			})

			Context("when the YAML cannot be parsed", func() {
				BeforeEach(func() {
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.ReplaceOperation, Path: "/applications/0/env/a", Was: "b", Value: "c"},
						},
					}
					rawManifest = []byte(`not-real-yaml!`)
				})

				It("returns an error", func() {
					Expect(err).To(MatchError("unable to process manifest diff because its format is invalid"))
				})
			})
		})

	})
})
