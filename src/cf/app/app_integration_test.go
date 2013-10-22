// Test only works on unix machines at the moment

// +build darwin freebsd linux netbsd openbsd

package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunningCommands(t *testing.T) {
	stdout, _, err := runCommand(t, "api")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "API endpoint")

	stdout, _, err = runCommand(t, "app")
	assert.Error(t, err)
	assert.Contains(t, stdout, "FAILED")

	stdout, _, err = runCommand(t, "target", "foo", "bar")
	assert.Error(t, err)
	assert.Contains(t, stdout, "FAILED")
}

func TestHelpCommand(t *testing.T) {
	helpOutput, _, err := runCommand(t, "help")
	assert.NoError(t, err)

	for _, cmdName := range availableCmdNames() {
		included := strings.Contains(helpOutput, "\n   "+cmdName)
		assert.True(t, included, "Could not find command %s in help text", cmdName)
	}
}

func runCommand(t *testing.T, params ...string) (stdout, stderr string, err error) {
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	sourceFile := filepath.Join(currentDir, "..", "..", "..", "src", "main", "cf.go")

	args := append([]string{"run", sourceFile}, params...)
	cmd := exec.Command("go", args...)

	stdoutWriter := bytes.NewBufferString("")
	stderrWriter := bytes.NewBufferString("")
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	err = cmd.Start()
	assert.NoError(t, err)

	err = cmd.Wait()
	stdout = string(stdoutWriter.Bytes())
	stderr = string(stderrWriter.Bytes())
	return
}
