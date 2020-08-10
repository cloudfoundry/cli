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

		Context("Operation kinds", func() {
			When("additions", func() {
				BeforeEach(func() {
					rawManifest = []byte("applications:\n- name: dora\n  env:\n    a: b\n    r: m")
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/0/env/r", Value: "m"},
						},
					}
				})

				It("outputs correctly formatted diff", func() {
					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    0:
      name: dora
      env:
        a: b
+       r: m
`))
				})
			})

			When("removal", func() {
				BeforeEach(func() {
					rawManifest = []byte("applications:\n- name: dora\n  env:\n    r: m")
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.RemoveOperation, Path: "/applications/0/env/a", Was: "b"},
						},
					}
				})

				It("outputs correctly formatted diff", func() {
					fmt.Printf("%+v", string(outputBuffer.Contents()))
					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    0:
      name: dora
      env:
-       a: b
        r: m
`))
				})
			})

			When("changes", func() {
				BeforeEach(func() {
					rawManifest = []byte("applications:\n- name: dora\n  env:\n    a: c\n    r: m")
					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.ReplaceOperation, Path: "/applications/0/env/a", Was: "b", Value: "c"},
						},
					}
				})

				It("outputs correctly formatted diff", func() {
					fmt.Printf("%+v", string(outputBuffer.Contents()))
					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    0:
      name: dora
      env:
-       a: b
+       a: c
        r: m
`))
				})
			})

		})

		Context("Edge Cases", func() {
			When("array", func() {
				BeforeEach(func() {
					rawManifest = []byte(`applications:
    -
      name: dora
      env:
        a: b
        r: m
    -
      name: dora1
`)

					diff = resources.ManifestDiff{
						Diffs: []resources.Diff{
							{Op: resources.AddOperation, Path: "/applications/1", Value: map[string]interface{}{"name": "dora1"}},
						},
					}
				})

				It("outputs correctly formatted diff for array", func() {
					fmt.Printf("%+v", string(outputBuffer.Contents()))
					Expect(string(outputBuffer.Contents())).To(Equal(`  ---
  applications:
    0:
      name: dora
      env:
        a: b
        r: m
+   1:
+     name: dora1
`))
				})
			})

			When("maps", func() {
			})

			When("multiline", func() {
			})
		})

		// When("example", func() {
		// 	BeforeEach(func() {
		// 		rawManifest = []byte("something")
		// 		diff = resources.ManifestDiff{}
		// 	})

		// 	It("does something", func() {
		// 		Expect(testUI.Out).To(Say("..."))
		// 	})
		// })
	})
})
