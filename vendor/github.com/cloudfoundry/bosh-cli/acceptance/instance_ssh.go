package acceptance

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type InstanceSSH interface {
	RunCommand(cmd string) (stdout, stderr string, exitCode int, err error)
	RunCommandWithSudo(cmd string) (stdout, stderr string, exitCode int, err error)
}

type instanceSSH struct {
	instanceUsername string
	instanceIP       string
	instancePassword string
	runner           boshsys.CmdRunner
	fileSystem       boshsys.FileSystem
}

func NewInstanceSSH(
	instanceUsername string,
	instanceIP string,
	instancePassword string,
	fileSystem boshsys.FileSystem,
	logger boshlog.Logger,
) InstanceSSH {
	return &instanceSSH{
		instanceUsername: instanceUsername,
		instanceIP:       instanceIP,
		instancePassword: instancePassword,
		runner:           boshsys.NewExecCmdRunner(logger),
		fileSystem:       fileSystem,
	}
}

func (s *instanceSSH) setupSSH() (boshsys.File, error) {
	sshConfigFile, err := s.fileSystem.TempFile("ssh-config")
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating temp ssh-config file")
	}

	success := false
	defer func() {
		if !success {
			_ = s.fileSystem.RemoveAll(sshConfigFile.Name())
		}
	}()

	sshConfigTemplate := `
Host warden-vm
	Hostname %s
	User %s
	StrictHostKeyChecking no
`
	sshConfig := fmt.Sprintf(
		sshConfigTemplate,
		s.instanceIP,
		s.instanceUsername,
	)

	err = s.fileSystem.WriteFileString(sshConfigFile.Name(), sshConfig)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Writing to temp ssh-config file: '%s'", sshConfigFile.Name())
	}

	success = true
	return sshConfigFile, nil
}

func (s *instanceSSH) RunCommand(cmd string) (stdout, stderr string, exitCode int, err error) {
	sshConfigFile, err := s.setupSSH()
	if err != nil {
		return "", "", -1, bosherr.WrapError(err, "Setting up SSH")
	}
	defer s.fileSystem.RemoveAll(sshConfigFile.Name())

	return s.runner.RunCommand(
		"sshpass",
		"-p"+s.instancePassword,
		"ssh",
		"warden-vm",
		"-F",
		sshConfigFile.Name(),
		cmd,
	)
}

func (s *instanceSSH) RunCommandWithSudo(cmd string) (stdout, stderr string, exitCode int, err error) {
	sshConfigFile, err := s.setupSSH()
	if err != nil {
		return "", "", -1, bosherr.WrapError(err, "Setting up SSH")
	}
	defer s.fileSystem.RemoveAll(sshConfigFile.Name())

	return s.runner.RunCommand(
		"sshpass",
		"-p"+s.instancePassword,
		"ssh",
		"warden-vm",
		"-F",
		sshConfigFile.Name(),
		fmt.Sprintf("echo %s | sudo -p '' -S %s", s.instancePassword, cmd),
	)
}
