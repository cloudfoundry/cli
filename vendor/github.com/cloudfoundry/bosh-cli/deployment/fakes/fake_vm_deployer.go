package fakes

import (
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type FakeVMDeployer struct {
	DeployInputs  []VMDeployInput
	DeployOutputs []vmDeployOutput

	WaitUntilReadyInputs []WaitUntilReadyInput
	WaitUntilReadyErr    error
}

type VMDeployInput struct {
	Cloud            bicloud.Cloud
	Manifest         bideplmanifest.Manifest
	Stemcell         bistemcell.CloudStemcell
	MbusURL          string
	EventLoggerStage biui.Stage
}

type WaitUntilReadyInput struct {
	VM               bivm.VM
	SSHTunnelOptions bisshtunnel.Options
	EventLoggerStage biui.Stage
}

type vmDeployOutput struct {
	vm  bivm.VM
	err error
}

func NewFakeVMDeployer() *FakeVMDeployer {
	return &FakeVMDeployer{
		DeployInputs:  []VMDeployInput{},
		DeployOutputs: []vmDeployOutput{},
	}
}

func (m *FakeVMDeployer) Deploy(
	cloud bicloud.Cloud,
	deploymentManifest bideplmanifest.Manifest,
	stemcell bistemcell.CloudStemcell,
	mbusURL string,
	eventLoggerStage biui.Stage,
) (bivm.VM, error) {
	input := VMDeployInput{
		Cloud:            cloud,
		Manifest:         deploymentManifest,
		Stemcell:         stemcell,
		MbusURL:          mbusURL,
		EventLoggerStage: eventLoggerStage,
	}
	m.DeployInputs = append(m.DeployInputs, input)

	output := m.DeployOutputs[0]
	m.DeployOutputs = m.DeployOutputs[1:]

	return output.vm, output.err
}

func (m *FakeVMDeployer) WaitUntilReady(vm bivm.VM, sshTunnelOptions bisshtunnel.Options, eventLoggerStage biui.Stage) error {
	input := WaitUntilReadyInput{
		VM:               vm,
		SSHTunnelOptions: sshTunnelOptions,
		EventLoggerStage: eventLoggerStage,
	}
	m.WaitUntilReadyInputs = append(m.WaitUntilReadyInputs, input)

	return m.WaitUntilReadyErr
}

func (m *FakeVMDeployer) SetDeployBehavior(vm bivm.VM, err error) {
	m.DeployOutputs = append(m.DeployOutputs, vmDeployOutput{vm: vm, err: err})
}
