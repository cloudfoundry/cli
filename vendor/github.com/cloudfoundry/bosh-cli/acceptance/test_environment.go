package acceptance

import (
	"fmt"
	"os"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type Environment interface {
	Home() string
	Path(string) string
	Copy(string, string) error
	WriteContent(string, []byte) error
}

type testEnvironment struct {
	cmdRunner  boshsys.CmdRunner
	fileSystem boshsys.FileSystem
	home       string
}

func NewTestEnvironment(fileSystem boshsys.FileSystem, logger boshlog.Logger) Environment {
	return testEnvironment{
		cmdRunner:  boshsys.NewExecCmdRunner(logger),
		fileSystem: fileSystem,
		home:       os.TempDir(),
	}
}

func (e testEnvironment) Home() string {
	return e.home
}

func (e testEnvironment) Path(name string) string {
	return filepath.Join(e.home, name)
}

func (e testEnvironment) Copy(destName, srcPath string) error {
	if srcPath == "" {
		return fmt.Errorf("Cannot use an empty source file path '' for destination file '%s'", destName)
	}

	_, _, exitCode, err := e.cmdRunner.RunCommand(
		"cp",
		srcPath,
		e.Path(destName),
	)

	if exitCode != 0 {
		return fmt.Errorf("cp of '%s' to '%s' failed", srcPath, destName)
	}

	return err
}

func (e testEnvironment) WriteContent(destName string, contents []byte) error {
	tmpFile, err := e.fileSystem.TempFile("bosh-cli-acceptance")
	if err != nil {
		return err
	}

	defer e.fileSystem.RemoveAll(tmpFile.Name())

	_, err = tmpFile.Write(contents)
	if err != nil {
		return err
	}

	err = tmpFile.Close()
	if err != nil {
		return err
	}

	return e.Copy(destName, tmpFile.Name())
}
