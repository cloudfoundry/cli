// Test only works on unix machines at the moment

// +build darwin freebsd linux netbsd openbsd

package app

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunningCommands(t *testing.T) {
	stdout, _, err := runCommand(t, "api")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "API endpoint")

	stdout, _, err = runCommand(t, "app")
	assert.Error(t, err)
	assert.Contains(t, stdout, "FAILED")

	stdout, _, err = runCommand(t, "h", "target")
	assert.NoError(t, err)
	assert.Contains(t, stdout, "TIP:")
	assert.Contains(t, stdout, "Use 'cf api' to set or view the target api url.")
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
