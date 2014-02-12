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
	m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"path": "/another/path-segment",
			},
		},
	}))

	Expect(errs).To(BeEmpty())
	Expect(*m.Applications[0].Path).To(Equal("/another/path-segment"))
}

func testManifestWithAbsolutePathOnWindows() {
	m, errs := manifest.NewManifest(`C:\some\path`, generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"path": `C:\another\path`,
			},
		},
	}))

	Expect(errs).To(BeEmpty())
	Expect(*m.Applications[0].Path).To(Equal(`C:\another\path`))
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestManifestWithGlobalAndAppSpecificProperties", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"instances": "3",
			"memory":    "512M",
			"applications": []interface{}{
				map[string]interface{}{
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
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"instances": "3",
			"memory":    "512",
			"applications": []interface{}{
				map[string]interface{}{
					"name": "bitcoin-miner",
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("memory"))
	})

	It("TestManifestWithTimeoutSetsHealthCheckTimeout", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{
					"name":    "bitcoin-miner",
					"timeout": "360",
				},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].HealthCheckTimeout).To(Equal(360))
	})

	It("TestManifestWithEmptyEnvVarIsInvalid", func() {
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"env": generic.NewMap(map[string]interface{}{
				"bar": nil,
			}),
			"applications": []interface{}{
				map[string]interface{}{
					"name": "bad app",
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("env var 'bar' should not be null"))
	})

	It("returns an empty map when no env was present in the manifest", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{"name": "no-env-vars"},
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
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{
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
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{
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

	It("TestParsingManifestWithPropertiesReturnsErrors", func() {
		_, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{
					"env": map[string]interface{}{
						"bar": "many-${foo}-are-cool",
					},
				},
			},
		}))

		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("Properties are not supported. Found property '${foo}'"))
	})

	It("TestParsingManifestWithNullCommand", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{
					"command": nil,
				},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(*m.Applications[0].Command).To(Equal(""))
	})

	It("TestParsingEmptyManifestDoesNotSetCommand", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
			"applications": []interface{}{
				map[string]interface{}{},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(m.Applications[0].Command).To(BeNil())
	})
})
