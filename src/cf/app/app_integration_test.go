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
	projectDir := filepath.Join(currentDir, "../../..")
	sourceFile := filepath.Join(projectDir, "src", "main", "cf.go")
	goFile := filepath.Join(projectDir, "bin", "go")

	args := append([]string{"run", sourceFile}, params...)
	goCmd := exec.Command(goFile, args...)

	stdoutWriter := bytes.NewBufferString("")
	stderrWriter := bytes.NewBufferString("")
	goCmd.Stdout = stdoutWriter
	goCmd.Stderr = stderrWriter

	err = goCmd.Start()
	assert.NoError(t, err)

	err = goCmd.Wait()

	return string(stdoutWriter.Bytes()), string(stderrWriter.Bytes()), err
}
