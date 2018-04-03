package instance

import (
	biblobstore "github.com/cloudfoundry/bosh-cli/blobstore"
	biinstancestate "github.com/cloudfoundry/bosh-cli/deployment/instance/state"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Factory interface {
	NewInstance(
		jobName string,
		id int,
		vm bivm.VM,
		vmManager bivm.Manager,
		sshTunnelFactory bisshtunnel.Factory,
		blobstore biblobstore.Blobstore,
		logger boshlog.Logger,
	) Instance
}

type factory struct {
	stateBuilderFactory biinstancestate.BuilderFactory
}

func NewFactory(
	stateBuilderFactory biinstancestate.BuilderFactory,
) Factory {
	return &factory{
		stateBuilderFactory: stateBuilderFactory,
	}
}

func (f *factory) NewInstance(
	jobName string,
	id int,
	vm bivm.VM,
	vmManager bivm.Manager,
	sshTunnelFactory bisshtunnel.Factory,
	blobstore biblobstore.Blobstore,
	logger boshlog.Logger,
) Instance {
	stateBuilder := f.stateBuilderFactory.NewBuilder(blobstore, vm.AgentClient())

	return NewInstance(
		jobName,
		id,
		vm,
		vmManager,
		sshTunnelFactory,
		stateBuilder,
		logger,
	)
}
