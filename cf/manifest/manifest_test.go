package manifest_test

import (
	"runtime"
	"strings"

	"code.cloudfoundry.org/cli/cf/manifest"
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
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
		Expect(*apps[0].Memory).To(Equal(int64(512)))
		Expect(apps[0].NoRoute).To(BeTrue())
	})

	Context("when there is no applications block", func() {
		It("returns a single application with the global properties", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"instances": "3",
				"memory":    "512M",
			}))

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(apps)).To(Equal(1))
			Expect(*apps[0].InstanceCount).To(Equal(3))
			Expect(*apps[0].Memory).To(Equal(int64(512)))
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
		Expect(err.Error()).To(ContainSubstring("Invalid value for 'memory': 512"))
	})

	It("returns an error when the memory limit is a non-string", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"instances": "3",
			"memory":    128,
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bitcoin-miner",
				},
			},
		}))

		_, err := m.Applications()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Invalid value for 'memory': 128"))
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

	It("allows boolean env var values", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"env": generic.NewMap(map[interface{}]interface{}{
				"bar": true,
			}),
		}))

		_, err := m.Applications()
		Expect(err).ToNot(HaveOccurred())
	})

	It("allows nil value for global env if env is present in the app", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"env": nil,
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bad app",
					"env": map[interface{}]interface{}{
						"foo": "bar",
					},
				},
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(*apps[0].EnvironmentVars).To(Equal(map[string]interface{}{"foo": "bar"}))
	})

	It("does not allow nil value for env in application", func() {
		m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
			"env": generic.NewMap(map[interface{}]interface{}{
				"foo": "bar",
			}),
			"applications": []interface{}{
				map[interface{}]interface{}{
					"name": "bad app",
					"env":  nil,
				},
			},
		}))

		_, err := m.Applications()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("env should not be null"))
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
					"no-hostname":  nil,
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
			"memory", "instances", "timeout", "no-route", "no-hostname", "services", "env", "random-route"}

		for _, key := range manifestKeys {
			Expect(errorSlice).To(ContainSubstrings([]string{key, "not be null"}))
		}
	})

	It("returns errors when hosts/domains is not valid slice", func() {
		m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"hosts":   "bad-value",
					"domains": []interface{}{"val1", "val2", false, true},
				},
			},
		}))

		_, err := m.Applications()
		Expect(err).To(HaveOccurred())
		errorSlice := strings.Split(err.Error(), "\n")

		Expect(errorSlice).To(ContainSubstrings([]string{"hosts", "to be a list of strings"}))
		Expect(errorSlice).To(ContainSubstrings([]string{"domains", "to be a list of strings"}))
	})

	It("parses known manifest keys", func() {
		m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"buildpack":         "my-buildpack",
					"disk_quota":        "512M",
					"domain":            "my-domain",
					"domains":           []interface{}{"domain1.test", "domain2.test"},
					"host":              "my-hostname",
					"hosts":             []interface{}{"host-1", "host-2"},
					"name":              "my-app-name",
					"stack":             "my-stack",
					"memory":            "256M",
					"health-check-type": "none",
					"instances":         1,
					"timeout":           11,
					"no-route":          true,
					"no-hostname":       true,
					"random-route":      true,
				},
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(apps)).To(Equal(1))

		Expect(*apps[0].BuildpackURL).To(Equal("my-buildpack"))
		Expect(*apps[0].DiskQuota).To(Equal(int64(512)))
		Expect(apps[0].Domains).To(ConsistOf([]string{"domain1.test", "domain2.test", "my-domain"}))
		Expect(apps[0].Hosts).To(ConsistOf([]string{"host-1", "host-2", "my-hostname"}))
		Expect(*apps[0].Name).To(Equal("my-app-name"))
		Expect(*apps[0].StackName).To(Equal("my-stack"))
		Expect(*apps[0].HealthCheckType).To(Equal("none"))
		Expect(*apps[0].Memory).To(Equal(int64(256)))
		Expect(*apps[0].InstanceCount).To(Equal(1))
		Expect(*apps[0].HealthCheckTimeout).To(Equal(11))
		Expect(apps[0].NoRoute).To(BeTrue())
		Expect(*apps[0].NoHostname).To(BeTrue())
		Expect(apps[0].UseRandomRoute).To(BeTrue())
	})

	It("removes duplicated values in 'hosts' and 'domains'", func() {
		m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
			"applications": []interface{}{
				map[interface{}]interface{}{
					"domain":  "my-domain",
					"domains": []interface{}{"my-domain", "domain1.test", "domain1.test", "domain2.test"},
					"host":    "my-hostname",
					"hosts":   []interface{}{"my-hostname", "host-1", "host-1", "host-2"},
					"name":    "my-app-name",
				},
			},
		}))

		apps, err := m.Applications()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(apps)).To(Equal(1))

		Expect(len(apps[0].Domains)).To(Equal(3))
		Expect(apps[0].Domains).To(ConsistOf([]string{"my-domain", "domain1.test", "domain2.test"}))
		Expect(len(apps[0].Hosts)).To(Equal(3))
		Expect(apps[0].Hosts).To(ConsistOf([]string{"my-hostname", "host-1", "host-2"}))
	})

	Context("old-style property syntax", func() {
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
		Expect(*apps[0].BuildpackURL).To(Equal(""))
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
		Expect(*apps[0].BuildpackURL).To(Equal(""))
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

	Context("parsing app ports", func() {
		It("parses app ports", func() {
			m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"app-ports": []interface{}{
							8080,
							9090,
						},
					},
				},
			}))

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect(apps[0].AppPorts).NotTo(BeNil())
			Expect(*(apps[0].AppPorts)).To(Equal([]int{8080, 9090}))
		})

		It("handles omitted field", func() {
			m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{},
				},
			}))

			apps, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect(apps[0].AppPorts).To(BeNil())
		})

		It("handles mixed arrays", func() {
			m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"app-ports": []interface{}{
							8080,
							"potato",
						},
					},
				},
			}))

			_, err := m.Applications()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected app-ports to be a list of integers."))
		})

		It("handles non-array values", func() {
			m := NewManifest("/some/path", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					map[interface{}]interface{}{
						"app-ports": "potato",
					},
				},
			}))

			_, err := m.Applications()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected app-ports to be a list of integers."))
		})
	})

	Context("parsing env vars", func() {
		It("handles values that are not strings", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"applications": []interface{}{
					generic.NewMap(map[interface{}]interface{}{
						"env": map[interface{}]interface{}{
							"string-key":      "value",
							"int-key":         1,
							"float-key":       11.1,
							"large-int-key":   123456789,
							"large-float-key": 123456789.12345678,
							"bool-key":        false,
						},
					}),
				},
			}))

			app, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect((*app[0].EnvironmentVars)["string-key"]).To(Equal("value"))
			Expect((*app[0].EnvironmentVars)["int-key"]).To(Equal("1"))
			Expect((*app[0].EnvironmentVars)["float-key"]).To(Equal("11.1"))
			Expect((*app[0].EnvironmentVars)["large-int-key"]).To(Equal("123456789"))
			Expect((*app[0].EnvironmentVars)["large-float-key"]).To(Equal("123456789.12345678"))
			Expect((*app[0].EnvironmentVars)["bool-key"]).To(Equal("false"))
		})
	})

	Context("parsing services", func() {
		It("can read a list of service instance names", func() {
			m := NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
				"services": []interface{}{"service-1", "service-2"},
			}))

			app, err := m.Applications()
			Expect(err).NotTo(HaveOccurred())

			Expect(app[0].ServicesToBind).To(Equal([]string{"service-1", "service-2"}))
		})
	})

	Context("when routes are provided", func() {
		var manifest *manifest.Manifest

		Context("when passed 'routes'", func() {
			Context("valid 'routes'", func() {
				BeforeEach(func() {
					manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
						"applications": []interface{}{
							generic.NewMap(map[interface{}]interface{}{
								"routes": []interface{}{
									map[interface{}]interface{}{"route": "route1.example.com"},
									map[interface{}]interface{}{"route": "route2.example.com"},
								},
							}),
						},
					}))
				})

				It("parses routes into app params", func() {
					apps, err := manifest.Applications()
					Expect(err).NotTo(HaveOccurred())
					Expect(apps).To(HaveLen(1))

					routes := apps[0].Routes
					Expect(routes).To(HaveLen(2))
					Expect(routes[0].Route).To(Equal("route1.example.com"))
					Expect(routes[1].Route).To(Equal("route2.example.com"))
				})
			})

			Context("invalid 'routes'", func() {
				Context("'routes' is formatted incorrectly", func() {
					BeforeEach(func() {
						manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
							"applications": []interface{}{
								generic.NewMap(map[interface{}]interface{}{
									"routes": []string{},
								}),
							},
						}))
					})

					It("errors out", func() {
						_, err := manifest.Applications()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(MatchRegexp("should be a list"))
					})
				})

				Context("an individual 'route' is formatted incorrectly", func() {
					BeforeEach(func() {
						manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
							"applications": []interface{}{
								generic.NewMap(map[interface{}]interface{}{
									"routes": []interface{}{
										map[interface{}]interface{}{"routef": "route1.example.com"},
									},
								}),
							},
						}))
					})

					It("parses routes into app params", func() {
						_, err := manifest.Applications()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(MatchRegexp("each route in 'routes' must have a 'route' property"))
					})
				})
			})
		})

		Context("when there are no routes", func() {
			BeforeEach(func() {
				manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
					"applications": []interface{}{
						generic.NewMap(map[interface{}]interface{}{
							"buildpack": nil,
							"command":   "echo banana",
						}),
					},
				}))
			})

			It("sets routes to be nil", func() {
				apps, err := manifest.Applications()
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(HaveLen(1))
				Expect(apps[0].Routes).To(BeNil())
			})
		})

		Context("when no-hostname is not specified in the manifest", func() {
			BeforeEach(func() {
				manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
					"applications": []interface{}{
						generic.NewMap(map[interface{}]interface{}{
							"buildpack": nil,
							"command":   "echo banana",
						}),
					},
				}))
			})

			It("sets no-hostname to be nil", func() {
				apps, err := manifest.Applications()
				Expect(err).NotTo(HaveOccurred())
				Expect(apps).To(HaveLen(1))
				Expect(apps[0].NoHostname).To(BeNil())
			})
		})

		Context("when no-hostname is specified in the manifest", func() {
			Context("and it is set to true", func() {
				Context("and the value is a boolean", func() {
					BeforeEach(func() {
						manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
							"applications": []interface{}{
								generic.NewMap(map[interface{}]interface{}{
									"buildpack":   nil,
									"command":     "echo banana",
									"no-hostname": true,
								}),
							},
						}))
					})

					It("sets no-hostname to be true", func() {
						apps, err := manifest.Applications()
						Expect(err).NotTo(HaveOccurred())
						Expect(apps).To(HaveLen(1))
						Expect(*apps[0].NoHostname).To(BeTrue())
					})
				})
				Context("and the value is a string", func() {
					BeforeEach(func() {
						manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
							"applications": []interface{}{
								generic.NewMap(map[interface{}]interface{}{
									"buildpack":   nil,
									"command":     "echo banana",
									"no-hostname": "true",
								}),
							},
						}))
					})

					It("sets no-hostname to be true", func() {
						apps, err := manifest.Applications()
						Expect(err).NotTo(HaveOccurred())
						Expect(apps).To(HaveLen(1))
						Expect(*apps[0].NoHostname).To(BeTrue())
					})
				})
			})
			Context("and it is set to false", func() {
				BeforeEach(func() {
					manifest = NewManifest("/some/path/manifest.yml", generic.NewMap(map[interface{}]interface{}{
						"applications": []interface{}{
							generic.NewMap(map[interface{}]interface{}{
								"buildpack":   nil,
								"command":     "echo banana",
								"no-hostname": false,
							}),
						},
					}))
				})
				It("sets no-hostname to be false", func() {
					apps, err := manifest.Applications()
					Expect(err).NotTo(HaveOccurred())
					Expect(apps).To(HaveLen(1))
					Expect(*apps[0].NoHostname).To(BeFalse())
				})
			})
		})
	})
})
