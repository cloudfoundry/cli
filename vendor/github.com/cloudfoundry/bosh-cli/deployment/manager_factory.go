package deployment

import (
	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	biinstance "github.com/cloudfoundry/bosh-cli/deployment/instance"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
)

type ManagerFactory interface {
	NewManager(bicloud.Cloud, biagentclient.AgentClient, biblobstore.Blobstore) Manager
}

type managerFactory struct {
	vmManagerFactory       bivm.ManagerFactory
	instanceManagerFactory biinstance.ManagerFactory
	diskManagerFactory     bidisk.ManagerFactory
	stemcellManagerFactory bistemcell.ManagerFactory
	deploymentFactory      Factory
}

func NewManagerFactory(
	vmManagerFactory bivm.ManagerFactory,
	instanceManagerFactory biinstance.ManagerFactory,
	diskManagerFactory bidisk.ManagerFactory,
	stemcellManagerFactory bistemcell.ManagerFactory,
	deploymentFactory Factory,
) ManagerFactory {
	return &managerFactory{
		vmManagerFactory:       vmManagerFactory,
		instanceManagerFactory: instanceManagerFactory,
		diskManagerFactory:     diskManagerFactory,
		stemcellManagerFactory: stemcellManagerFactory,
		deploymentFactory:      deploymentFactory,
	}
}

func (f *managerFactory) NewManager(cloud bicloud.Cloud, agentClient biagentclient.AgentClient, blobstore biblobstore.Blobstore) Manager {
	vmManager := f.vmManagerFactory.NewManager(cloud, agentClient)
	instanceManager := f.instanceManagerFactory.NewManager(cloud, vmManager, blobstore)
	diskManager := f.diskManagerFactory.NewManager(cloud)
	stemcellManager := f.stemcellManagerFactory.NewManager(cloud)

	return NewManager(instanceManager, diskManager, stemcellManager, f.deploymentFactory)
}
