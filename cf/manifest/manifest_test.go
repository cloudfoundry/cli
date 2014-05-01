package manifest_test

import (
	"github.com/cloudfoundry/cli/cf/manifest"
	"github.com/cloudfoundry/cli/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"runtime"
	"strings"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
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

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())

		Expect(*apps[0].InstanceCount).To(Equal(3))
		Expect(*apps[0].Memory).To(Equal(uint64(512)))
		Expect(apps[0].NoRoute).To(BeTrue())
	})

	Describe("when there is no applications block", func() {
		It("returns a single application with the global properties", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"instances": "3",
				"memory":    "512M",
			}))

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(apps)).To(Equal(1))
			Expect(*apps[0].InstanceCount).To(Equal(3))
			Expect(*apps[0].Memory).To(Equal(uint64(512)))
		})
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

		_, err := m.Applications()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("memory"))
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

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
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

		_, err := m.Applications()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("env var 'bar' should not be null"))
	})

	It("returns an empty map when no env was present in the manifest", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{"name": "no-env-vars"},
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
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

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect(*apps[0].Path).To(Equal(`C:\another\path`))
		} else {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"path": "/another/path-segment",
					},
				},
			}))

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
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

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
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
					"disk_quota":   nil,
					"domain":       nil,
					"host":         nil,
					"name":         nil,
					"path":         nil,
					"stack":        nil,
					"memory":       nil,
					"instances":    nil,
					"timeout":      nil,
					"no-route":     nil,
					"services":     nil,
					"env":          nil,
					"random-route": nil,
				},
			},
		}))

		_, err := m.Applications()
		Expect(err).To(HaveOccurred())
		errorSlice := strings.Split(err.Error(), "\n")
		manifestKeys := []string{"disk_quota", "domain", "host", "name", "path", "stack",
			"memory", "instances", "timeout", "no-route", "services", "env", "random-route"}

		for _, key := range manifestKeys {
			Expect(errorSlice).To(ContainSubstrings([]string{key, "not be null"}))
		}
	})

	It("parses known manifest keys", func() {
		m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"buildpack":    "my-buildpack",
					"disk_quota":   "512M",
					"domain":       "my-domain",
					"host":         "my-hostname",
					"name":         "my-app-name",
					"stack":        "my-stack",
					"memory":       "256M",
					"instances":    1,
					"timeout":      11,
					"no-route":     true,
					"random-route": true,
				},
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(apps)).To(Equal(1))

		Expect(*apps[0].BuildpackUrl).To(Equal("my-buildpack"))
		Expect(*apps[0].DiskQuota).To(Equal(uint64(512)))
		Expect(*apps[0].Domain).To(Equal("my-domain"))
		Expect(*apps[0].Host).To(Equal("my-hostname"))
		Expect(*apps[0].Name).To(Equal("my-app-name"))
		Expect(*apps[0].StackName).To(Equal("my-stack"))
		Expect(*apps[0].Memory).To(Equal(uint64(256)))
		Expect(*apps[0].InstanceCount).To(Equal(1))
		Expect(*apps[0].HealthCheckTimeout).To(Equal(11))
		Expect(apps[0].NoRoute).To(BeTrue())
		Expect(apps[0].UseRandomHostname).To(BeTrue())
	})

	Describe("old-style property syntax", func() {
		It("returns an error when the manifest contains non-whitelist properties", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					generic.NewMap(map[interface{}]interface{}{
						"env": generic.NewMap(map[interface{}]interface{}{
							"bar": "many-${some_property-name}-are-cool",
						}),
					}),
				},
			}))

			_, err := m.Applications()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("'${some_property-name}'"))
		})

		It("replaces the '${random-word} with a combination of 2 random words", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					generic.NewMap(map[interface{}]interface{}{
						"env": generic.NewMap(map[interface{}]interface{}{
							"bar": "prefix_${random-word}_suffix",
							"foo": "some-value",
						}),
					}),
				},
			}))

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())
			Expect((*apps[0].EnvironmentVars)["bar"]).To(MatchRegexp(`prefix_\w+-\w+_suffix`))
			Expect((*apps[0].EnvironmentVars)["foo"]).To(Equal("some-value"))

			apps2, _ := m.Applications()
			Expect((*apps2[0].EnvironmentVars)["bar"]).To(MatchRegexp(`prefix_\w+-\w+_suffix`))
			Expect((*apps2[0].EnvironmentVars)["bar"]).NotTo(Equal((*apps[0].EnvironmentVars)["bar"]))
		})
	})

	It("sets the command and buildpack to blank when their values are null in the manifest", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"buildpack": nil,
					"command":   nil,
				}),
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(*apps[0].Command).To(Equal(""))
		Expect(*apps[0].BuildpackUrl).To(Equal(""))
	})

	It("sets the command and buildpack to blank when their values are 'default' in the manifest", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				generic.NewMap(map[interface{}]interface{}{
					"command":   "default",
					"buildpack": "default",
				}),
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(*apps[0].Command).To(Equal(""))
		Expect(*apps[0].BuildpackUrl).To(Equal(""))
	})

	It("does not set the start command when the manifest doesn't have the 'command' key", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{},
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
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

		apps1, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())

		apps2, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(apps1).To(Equal(apps2))
	})

	Describe("parsing env vars", func() {
		It("handles values that are not strings", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					generic.NewMap(map[interface{}]interface{}{
						"env": map[interface{}]interface{}{
							"string-key": "value",
							"int-key":    1,
							"float-key":  11.1,
						},
					}),
				},
			}))

			app, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect((*app[0].EnvironmentVars)["string-key"]).To(Equal("value"))
			Expect((*app[0].EnvironmentVars)["int-key"]).To(Equal("1"))
			Expect((*app[0].EnvironmentVars)["float-key"]).To(ContainSubstring("11.1"))
		})

		It("handles values that cannot be converted to strings", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					generic.NewMap(map[interface{}]interface{}{
						"env": map[interface{}]interface{}{
							"bad-key": map[interface{}]interface{}{},
						},
					}),
				},
			}))

			_, err := m.Applications()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("parsing services", func() {
		It("can read a list of service instance names", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"services": []interface{}{"service-1", "service-2"},
			}))

			app, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect(*app[0].ServicesToBind).To(Equal([]string{"service-1", "service-2"}))
		})
	})
})
