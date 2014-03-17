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

func NewManifest(path string, data generic.Map) (m *manifest.Manifest) {
	return &manifest.Manifest{Path: path, Data: data}
}

var _ = Describe("Manifests", func() {
	It("merges global properties into each app's properties", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"instances": "3",
			"memory":    "512M",
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name":     "bitcoin-miner",
					"no-route": true,
				},
			},
		}))

		apps, errs := m.Applications()
		Expect(errs).To(BeEmpty())

		Expect(*apps[0].InstanceCount).To(Equal(3))
		Expect(*apps[0].Memory).To(Equal(uint64(512)))
		Expect(*apps[0].NoRoute).To(BeTrue())
	})

	It("returns an error when the memory limit doesn't have a unit", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"instances": "3",
			"memory":    "512",
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bitcoin-miner",
				},
			},
		}))

		_, errs := m.Applications()
		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("memory"))
	})

	It("sets applications' health check timeouts", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name":    "bitcoin-miner",
					"timeout": "360",
				},
			},
		}))

		apps, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(*apps[0].HealthCheckTimeout).To(Equal(360))
	})

	It("does not allow nil values for environment variables", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"env": generic.NewMap(map[interface{}]interface{}{
				"bar": nil,
			}),
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bad app",
				},
			},
		}))

		_, errs := m.Applications()
		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("env var 'bar' should not be null"))
	})

	It("returns an empty map when no env was present in the manifest", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{"name": "no-env-vars"},
			},
		}))

		apps, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(*apps[0].EnvironmentVars).NotTo(BeNil())
	})

	It("allows applications to have absolute paths", func() {
		if runtime.GOOS == "windows" {
			m := NewManifest(`C:\some\path\manifest.yml`, generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"path": `C:\another\path`,
					},
				},
			}))

			apps, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(*apps[0].Path).To(Equal(`C:\another\path`))
		} else {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"path": "/another/path-segment",
					},
				},
			}))

			apps, errs := m.Applications()
			Expect(errs).To(BeEmpty())
			Expect(*apps[0].Path).To(Equal("/another/path-segment"))
		}
	})

	It("expands relative app paths based on the manifest's path", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"path": "../another/path-segment",
				},
			},
		}))

		apps, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		if runtime.GOOS == "windows" {
			Expect(*apps[0].Path).To(Equal("\\some\\another\\path-segment"))
		} else {
			Expect(*apps[0].Path).To(Equal("/some/another/path-segment"))
		}
	})

	It("returns errors when there are null values", func() {
		m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
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

		_, errs := m.Applications()
		Expect(errs).NotTo(BeEmpty())
		errorSlice := strings.Split(errs.Error(), "\n")
		manifestKeys := []string{"buildpack", "disk_quota", "domain", "host", "name", "path", "stack",
			"memory", "instances", "timeout", "no-route", "services", "env"}

		for _, key := range manifestKeys {
			testassert.SliceContains(errorSlice, testassert.Lines{{key, "not be null"}})
		}
	})

	It("returns an error when the manifest contains old-style property syntax", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"env": map[interface{}]interface{}{
						"bar": "many-${some_property-name}-are-cool",
					},
				},
			},
		}))

		_, errs := m.Applications()
		Expect(errs).NotTo(BeEmpty())
		Expect(errs.Error()).To(ContainSubstring("'${some_property-name}'"))
	})

	It("sets the command to blank when its value is null in the manifest", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"command": nil,
				},
			},
		}))

		apps, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(*apps[0].Command).To(Equal(""))
	})

	It("does not set the start command when the manifest doesn't have the 'command' key", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{},
			},
		}))

		apps, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(apps[0].Command).To(BeNil())
	})

	It("can build the applications multiple times", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"memory": "254m",
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bitcoin-miner",
				},
				map[interface{}]interface{}{
					"name": "bitcoin-miner",
				},
			},
		}))

		apps1, errs := m.Applications()
		apps2, errs := m.Applications()
		Expect(errs).To(BeEmpty())
		Expect(apps1).To(Equal(apps2))
	})
})
