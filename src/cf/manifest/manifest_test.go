package manifest_test

import (
	"cf/manifest"
	"generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"runtime"
	"strings"
	testassert "testhelpers/assert"
)

func testManifestWithAbsolutePathOnPosix() {
	m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
		"applications": []interface{}{
			map[interface{}]interface{}{
				"path": "/another/path-segment",
			},
		},
	}))

	Expect(errs).To(BeEmpty())
	Expect(*m.Applications[0].Path).To(Equal("/another/path-segment"))
}

func testManifestWithAbsolutePathOnWindows() {
	m, errs := manifest.NewManifest(`C:\some\path`, generic.NewMap(map[interface{}]interface{}{
		"applications": []interface{}{
			map[interface{}]interface{}{
				"path": `C:\another\path`,
			},
		},
	}))

	Expect(errs).To(BeEmpty())
	Expect(*m.Applications[0].Path).To(Equal(`C:\another\path`))
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestManifestWithGlobalAndAppSpecificProperties", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"instances": "3",
			"memory":    "512M",
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name":     "bitcoin-miner",
					"no-route": true,
				},
			},
		}))

		Expect(errs).To(BeEmpty())

		apps := m.Applications
		Expect(*apps[0].InstanceCount).To(Equal(3))
		Expect(*apps[0].Memory).To(Equal(uint64(512)))
		Expect(*apps[0].NoRoute).To(BeTrue())
	})

	It("TestManifestWithInvalidMemory", func() {
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"instances": "3",
			"memory":    "512",
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bitcoin-miner",
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("memory"))
	})

	It("TestManifestWithTimeoutSetsHealthCheckTimeout", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name":    "bitcoin-miner",
					"timeout": "360",
				},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].HealthCheckTimeout).To(Equal(360))
	})

	It("TestManifestWithEmptyEnvVarIsInvalid", func() {
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"env": generic.NewMap(map[interface{}]interface{}{
				"bar": nil,
			}),
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bad app",
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("env var 'bar' should not be null"))
	})

	It("returns an empty map when no env was present in the manifest", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{"name": "no-env-vars"},
			},
		}))
		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].EnvironmentVars).NotTo(BeNil())
	})

	It("TestManifestWithAbsolutePath", func() {
		if runtime.GOOS == "windows" {
			testManifestWithAbsolutePathOnWindows()
		} else {
			testManifestWithAbsolutePathOnPosix()
		}
	})

	It("TestManifestWithRelativePath", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"path": "../another/path-segment",
				},
			},
		}))

		Expect(errs).To(BeEmpty())
		if runtime.GOOS == "windows" {
			Expect(*m.Applications[0].Path).To(Equal("\\some\\another\\path-segment"))
		} else {
			Expect(*m.Applications[0].Path).To(Equal("/some/another/path-segment"))
		}
	})

	It("TestParsingManifestWithNulls", func() {
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"buildpack":  nil,
					"disk_quota": nil,
					"domain":     nil,
					"host":       nil,
					"name":       nil,
					"path":       nil,
					"stack":      nil,
					"memory":     nil,
					"instances":  nil,
					"timeout":    nil,
					"no-route":   nil,
					"services":   nil,
					"env":        nil,
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		errorSlice := strings.Split(errs.Error(), "\n")
		manifestKeys := []string{"buildpack", "disk_quota", "domain", "host", "name", "path", "stack",
			"memory", "instances", "timeout", "no-route", "services", "env"}

		for _, key := range manifestKeys {
			testassert.SliceContains(errorSlice, testassert.Lines{{key, "not be null"}})
		}
	})

	It("returns an error when the manifest contains old-style property syntax", func() {
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"env": map[interface{}]interface{}{
						"bar": "many-${some_property-name}-are-cool",
					},
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("'${some_property-name}'"))
	})

	It("TestParsingManifestWithNullCommand", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"command": nil,
				},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].Command).To(Equal(""))
	})

	It("TestParsingEmptyManifestDoesNotSetCommand", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(m.Applications[0].Command).To(BeNil())
	})
})
