package vm

import (
	"code.cloudfoundry.org/clock"
	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	biconfig "github.com/cloudfoundry/bosh-cli/config"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type ManagerFactory interface {
	NewManager(cloud bicloud.Cloud, agentClient biagentclient.AgentClient) Manager
}

type managerFactory struct {
	vmRepo        biconfig.VMRepo
	stemcellRepo  biconfig.StemcellRepo
	diskDeployer  DiskDeployer
	uuidGenerator boshuuid.Generator
	fs            boshsys.FileSystem
	logger        boshlog.Logger
}

func NewManagerFactory(
	vmRepo biconfig.VMRepo,
	stemcellRepo biconfig.StemcellRepo,
	diskDeployer DiskDeployer,
	uuidGenerator boshuuid.Generator,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) ManagerFactory {
	return &managerFactory{
		vmRepo:        vmRepo,
		stemcellRepo:  stemcellRepo,
		diskDeployer:  diskDeployer,
		uuidGenerator: uuidGenerator,
		fs:            fs,
		logger:        logger,
	}
}

func (f *managerFactory) NewManager(cloud bicloud.Cloud, agentClient biagentclient.AgentClient) Manager {
	return NewManager(
		f.vmRepo,
		f.stemcellRepo,
		f.diskDeployer,
		agentClient,
		cloud,
		f.uuidGenerator,
		f.fs,
		f.logger,
		clock.NewClock(),
	)
}
