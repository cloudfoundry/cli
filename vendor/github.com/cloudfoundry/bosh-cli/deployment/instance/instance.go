package instance

import (
	"fmt"
	"time"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	biinstancestate "github.com/cloudfoundry/bosh-cli/deployment/instance/state"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	bisshtunnel "github.com/cloudfoundry/bosh-cli/deployment/sshtunnel"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type Instance interface {
	JobName() string
	ID() int
	Disks() ([]bidisk.Disk, error)
	WaitUntilReady(biinstallmanifest.Registry, biui.Stage) error
	UpdateDisks(bideplmanifest.Manifest, biui.Stage) ([]bidisk.Disk, error)
	UpdateJobs(bideplmanifest.Manifest, biui.Stage) error
	Delete(
		pingTimeout time.Duration,
		pingDelay time.Duration,
		stage biui.Stage,
	) error
}

type instance struct {
	jobName          string
	id               int
	vm               bivm.VM
	vmManager        bivm.Manager
	sshTunnelFactory bisshtunnel.Factory
	stateBuilder     biinstancestate.Builder
	logger           boshlog.Logger
	logTag           string
}

func NewInstance(
	jobName string,
	id int,
	vm bivm.VM,
	vmManager bivm.Manager,
	sshTunnelFactory bisshtunnel.Factory,
	stateBuilder biinstancestate.Builder,
	logger boshlog.Logger,
) Instance {
	return &instance{
		jobName:          jobName,
		id:               id,
		vm:               vm,
		vmManager:        vmManager,
		sshTunnelFactory: sshTunnelFactory,
		stateBuilder:     stateBuilder,
		logger:           logger,
		logTag:           "instance",
	}
}

func (i *instance) JobName() string {
	return i.jobName
}

func (i *instance) ID() int {
	return i.id
}

func (i *instance) Disks() ([]bidisk.Disk, error) {
	disks, err := i.vm.Disks()
	if err != nil {
		return disks, bosherr.WrapError(err, "Listing instance disks")
	}
	return disks, nil
}

func (i *instance) WaitUntilReady(
	registryConfig biinstallmanifest.Registry,
	stage biui.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s' to be ready", i.vm.CID())
	err := stage.Perform(stepName, func() error {
		if !registryConfig.IsEmpty() {
			sshReadyErrCh := make(chan error)
			sshErrCh := make(chan error)

			sshTunnelOptions := bisshtunnel.Options{
				Host:              registryConfig.SSHTunnel.Host,
				Port:              registryConfig.SSHTunnel.Port,
				User:              registryConfig.SSHTunnel.User,
				Password:          registryConfig.SSHTunnel.Password,
				PrivateKey:        registryConfig.SSHTunnel.PrivateKey,
				LocalForwardPort:  registryConfig.Port,
				RemoteForwardPort: registryConfig.Port,
			}
			sshTunnel := i.sshTunnelFactory.NewSSHTunnel(sshTunnelOptions)
			go sshTunnel.Start(sshReadyErrCh, sshErrCh)

			err := <-sshReadyErrCh
			if err != nil {
				return bosherr.WrapError(err, "Starting SSH tunnel")
			}
		}

		return i.vm.WaitUntilReady(10*time.Minute, 500*time.Millisecond)
	})

	return err
}

func (i *instance) UpdateDisks(deploymentManifest bideplmanifest.Manifest, stage biui.Stage) ([]bidisk.Disk, error) {
	diskPool, err := deploymentManifest.DiskPool(i.jobName)
	if err != nil {
		return []bidisk.Disk{}, bosherr.WrapError(err, "Getting disk pool")
	}

	disks, err := i.vm.UpdateDisks(diskPool, stage)
	if err != nil {
		return disks, bosherr.WrapError(err, "Updating disks")
	}

	return disks, nil
}

func (i *instance) UpdateJobs(
	deploymentManifest bideplmanifest.Manifest,
	stage biui.Stage,
) error {
	initialAgentState, err := i.stateBuilder.BuildInitialState(i.jobName, i.id, deploymentManifest)
	if err != nil {
		return bosherr.WrapErrorf(err, "Building initial state for instance '%s/%d'", i.jobName, i.id)
	}

	// apply it to agent to force it to load networking details
	err = i.vm.Apply(initialAgentState.ToApplySpec())
	if err != nil {
		return bosherr.WrapError(err, "Applying the initial agent state")
	}

	// now that the agent will tell us the address, get new state
	resolvedAgentState, err := i.vm.GetState()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting state for instance '%s/%d'", i.jobName, i.id)
	}

	newAgentState, err := i.stateBuilder.Build(i.jobName, i.id, deploymentManifest, stage, resolvedAgentState)
	if err != nil {
		return bosherr.WrapErrorf(err, "Building state for instance '%s/%d'", i.jobName, i.id)
	}
	stepName := fmt.Sprintf("Updating instance '%s/%d'", i.jobName, i.id)
	err = stage.Perform(stepName, func() error {
		err = i.vm.Stop()
		if err != nil {
			return bosherr.WrapError(err, "Stopping the agent")
		}

		err = i.vm.Apply(newAgentState.ToApplySpec())
		if err != nil {
			return bosherr.WrapError(err, "Applying the agent state")
		}

		err = i.vm.RunScript("pre-start", map[string]interface{}{})
		if err != nil {
			return bosherr.WrapError(err, "Running the pre-start script")
		}

		err = i.vm.Start()
		if err != nil {
			return bosherr.WrapError(err, "Starting the agent")
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = i.waitUntilJobsAreRunning(deploymentManifest.Update.UpdateWatchTime, stage)
	if err != nil {
		return err
	}

	stepName = fmt.Sprintf("Running the post-start scripts '%s/%d'", i.jobName, i.id)
	err = stage.Perform(stepName, func() error {
		err = i.vm.RunScript("post-start", map[string]interface{}{})
		if err != nil {
			return bosherr.WrapError(err, "Running the post-start script")
		}
		return nil
	})

	return err
}

func (i *instance) Delete(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	stage biui.Stage,
) error {
	vmExists, err := i.vm.Exists()
	if err != nil {
		return bosherr.WrapErrorf(err, "Checking existence of vm for instance '%s/%d'", i.jobName, i.id)
	}

	if vmExists {
		if err = i.shutdown(pingTimeout, pingDelay, stage); err != nil {
			return err
		}
	}

	// non-existent VMs still need to be 'deleted' to clean up related resources owned by the CPI
	stepName := fmt.Sprintf("Deleting VM '%s'", i.vm.CID())
	return stage.Perform(stepName, func() error {
		err := i.vm.Delete()
		cloudErr, ok := err.(bicloud.Error)
		if ok && cloudErr.Type() == bicloud.VMNotFoundError {
			return biui.NewSkipStageError(cloudErr, "VM not found")
		}
		return err
	})
}

func (i *instance) shutdown(
	pingTimeout time.Duration,
	pingDelay time.Duration,
	stage biui.Stage,
) error {
	stepName := fmt.Sprintf("Waiting for the agent on VM '%s'", i.vm.CID())
	waitingForAgentErr := stage.Perform(stepName, func() error {
		if err := i.vm.WaitUntilReady(pingTimeout, pingDelay); err != nil {
			return bosherr.WrapError(err, "Agent unreachable")
		}
		return nil
	})
	if waitingForAgentErr != nil {
		i.logger.Warn(i.logTag, "Gave up waiting for agent: %s", waitingForAgentErr.Error())
		return nil
	}

	if err := i.stopJobs(stage); err != nil {
		return err
	}
	if err := i.unmountDisks(stage); err != nil {
		return err
	}
	return nil
}

func (i *instance) waitUntilJobsAreRunning(updateWatchTime bideplmanifest.WatchTime, stage biui.Stage) error {
	start := time.Duration(updateWatchTime.Start) * time.Millisecond
	end := time.Duration(updateWatchTime.End) * time.Millisecond
	delayBetweenAttempts := 1 * time.Second
	maxAttempts := int((end - start) / delayBetweenAttempts)

	stepName := fmt.Sprintf("Waiting for instance '%s/%d' to be running", i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		time.Sleep(start)
		return i.vm.WaitToBeRunning(maxAttempts, delayBetweenAttempts)
	})
}

func (i *instance) stopJobs(stage biui.Stage) error {
	stepName := fmt.Sprintf("Stopping jobs on instance '%s/%d'", i.jobName, i.id)
	return stage.Perform(stepName, func() error {
		return i.vm.Stop()
	})
}

func (i *instance) unmountDisks(stage biui.Stage) error {
	disks, err := i.vm.Disks()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting VM '%s' disks", i.vm.CID())
	}

	for _, disk := range disks {
		stepName := fmt.Sprintf("Unmounting disk '%s'", disk.CID())
		err = stage.Perform(stepName, func() error {
			if err := i.vm.UnmountDisk(disk); err != nil {
				return bosherr.WrapErrorf(err, "Unmounting disk '%s' from VM '%s'", disk.CID(), i.vm.CID())
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
