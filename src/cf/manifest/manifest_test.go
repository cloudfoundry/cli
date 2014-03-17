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

var _ = Describe("Manifests", func() {
	It("merges global properties into each app's properties", func() {
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

	It("returns an error when the memory limit doesn't have a unit", func() {
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

	It("sets applications' health check timeouts", func() {
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

	It("does not allow nil values for environment variables", func() {
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

	It("allows applications to have absolute paths", func() {
		if runtime.GOOS == "windows" {
			m, errs := manifest.NewManifest(`C:\some\path`, generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"path": `C:\another\path`,
					},
				},
			}))

			Expect(errs).To(BeEmpty())
			Expect(*m.Applications[0].Path).To(Equal(`C:\another\path`))
		} else {
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
	})

	It("expands relative app paths based on the manifest's path", func() {
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

	It("returns errors when there are null values", func() {
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

	It("sets the command to blank when its value is null in the manifest", func() {
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

	It("does not set the start command when the manifest doesn't have the 'command' key", func() {
		m, errs := manifest.NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{},
			},
		}))

		Expect(errs).To(BeEmpty())
		Expect(m.Applications[0].Command).To(BeNil())
	})
})
