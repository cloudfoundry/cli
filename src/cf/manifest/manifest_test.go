package manifest_test

import (
	"cf/manifest"
	"generic"
	"github.com/stretchr/testify/assert"
	"runtime"
	"strings"
	testassert "testhelpers/assert"
	"testing"
)

func TestManifestWithGlobalAndAppSpecificProperties(t *testing.T) {
	m, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"instances": "3",
		"memory":    "512M",
		"applications": []interface{}{
			map[string]interface{}{
				"name":     "bitcoin-miner",
				"no-route": true,
			},
		},
	}))
	assert.NoError(t, err)

	apps := m.Applications
	assert.Equal(t, apps[0].Get("instances"), 3)
	assert.Equal(t, apps[0].Get("memory").(uint64), uint64(512))
	assert.True(t, apps[0].Get("no-route").(bool))
}

func TestManifestWithInvalidMemory(t *testing.T) {
	_, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"instances": "3",
		"memory":    "512",
		"applications": []interface{}{
			map[string]interface{}{
				"name": "bitcoin-miner",
			},
		},
	}))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "memory")
}

func TestManifestWithTimeoutSetsHealthCheckTimeout(t *testing.T) {
	m, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"name":    "bitcoin-miner",
				"timeout": "360",
			},
		},
	}))

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("health_check_timeout"), 360)
	assert.False(t, m.Applications[0].Has("timeout"))
}

func TestManifestWithEmptyEnvVarIsInvalid(t *testing.T) {
	_, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"env": map[string]interface{}{
			"bar": nil,
		},
		"applications": []interface{}{
			map[string]interface{}{
				"name": "bad app",
			},
		},
	}))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "env var 'bar' should not be null")
}

func TestManifestWithAbsolutePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		testManifestWithAbsolutePathOnWindows(t)
	} else {
		testManifestWithAbsolutePathOnPosix(t)
	}
}

func testManifestWithAbsolutePathOnPosix(t *testing.T) {
	m, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"path": "/another/path-segment",
			},
		},
	}))

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("path"), "/another/path-segment")
}

func testManifestWithAbsolutePathOnWindows(t *testing.T) {
	m, err := manifest.NewManifest(`C:\some\path`, generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"path": `C:\another\path`,
			},
		},
	}))

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("path"), `C:\another\path`)
}

func TestManifestWithRelativePath(t *testing.T) {
	m, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"path": "../another/path-segment",
			},
		},
	}))

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("path"), "/some/another/path-segment")
}

func TestParsingManifestWithNulls(t *testing.T) {
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

	assert.Error(t, errs)
	errorSlice := strings.Split(errs.Error(), "\n")
	manifestKeys := []string{"buildpack", "disk_quota", "domain", "host", "name", "path", "stack",
		"memory", "instances", "timeout", "no-route", "services", "env"}

	for _, key := range manifestKeys {
		testassert.SliceContains(t, errorSlice, testassert.Lines{{key, "not be null"}})
	}
}

func TestParsingManifestWithPropertiesReturnsErrors(t *testing.T) {
	_, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"env": map[string]interface{}{
					"bar": "many-${foo}-are-cool",
				},
			},
		},
	}))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Properties are not supported. Found property '${foo}'")
}

func TestParsingManifestWithNullCommand(t *testing.T) {
	m, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{
				"command": nil,
			},
		},
	}))

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("command"), "")
}

func TestParsingEmptyManifestDoesNotSetCommand(t *testing.T) {
	m, err := manifest.NewManifest("/some/path", generic.NewMap(map[string]interface{}{
		"applications": []interface{}{
			map[string]interface{}{},
		},
	}))

	assert.NoError(t, err)
	assert.False(t, m.Applications[0].Has("command"))
}
