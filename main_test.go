package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"code.cloudfoundry.org/cli/v8/util/configv3"
)

// The test executable doubles as a plugin subprocess for the built CLI.
func TestMain(m *testing.M) {
	switch os.Getenv("CF_CLI_MAIN_TEST_PLUGIN") {
	case "fail":
		fmt.Fprintln(os.Stderr, "plugin diagnostic")
		os.Exit(3)
	case "silent-fail":
		os.Exit(3)
	case "success":
		fmt.Fprintln(os.Stdout, "plugin output")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestPluginProcessExit(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "cf")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("building CLI: %v\n%s", err, output)
	}
	pluginPath, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	for _, mode := range []string{"fail", "silent-fail", "success", "missing"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			pluginHome := filepath.Join(home, ".cf", "plugins")
			if err := os.MkdirAll(pluginHome, 0700); err != nil {
				t.Fatal(err)
			}
			location := pluginPath
			if mode == "missing" {
				location = filepath.Join(home, "nonexistent-plugin")
			}
			writeJSON := func(path string, value interface{}) {
				t.Helper()
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			writeJSON(filepath.Join(home, ".cf", "config.json"), configv3.JSONConfig{ConfigVersion: configv3.CurrentConfigVersion})
			writeJSON(filepath.Join(pluginHome, "config.json"), configv3.PluginsConfig{
				Plugins: map[string]configv3.Plugin{
					"test-plugin": {Location: location, Commands: []configv3.PluginCommand{{Name: "r07-plugin"}}},
				},
			})
			cmd := exec.Command(binary, "r07-plugin")
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, "HOME=") && !strings.HasPrefix(value, "CF_HOME=") && !strings.HasPrefix(value, "CF_PLUGIN_HOME=") && !strings.HasPrefix(value, "CF_CLI_MAIN_TEST_PLUGIN=") {
					cmd.Env = append(cmd.Env, value)
				}
			}
			cmd.Env = append(cmd.Env, "HOME="+home, "CF_CLI_MAIN_TEST_PLUGIN="+mode)
			if runtime.GOOS == "windows" {
				cmd.Env = append(cmd.Env, "HOMEDRIVE=", "HOMEPATH=", "USERPROFILE="+home)
			}
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if mode == "success" {
				if err != nil || stdout.String() != "plugin output\n" || stderr.Len() != 0 {
					t.Fatalf("success: err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
				}
				return
			}
			exitErr, ok := err.(*exec.ExitError)
			if !ok || exitErr.ExitCode() != 1 {
				t.Fatalf("expected CLI exit 1, got %v", err)
			}
			if mode == "missing" {
				if !strings.Contains(stdout.String()+stderr.String(), "nonexistent-plugin") {
					t.Fatalf("missing executable error not reported: stdout=%q stderr=%q", stdout.String(), stderr.String())
				}
				return
			}
			wantError := ""
			if mode == "fail" {
				wantError = "plugin diagnostic\n"
			}
			if stdout.Len() != 0 || stderr.String() != wantError {
				t.Fatalf("unexpected CLI diagnostics: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}
