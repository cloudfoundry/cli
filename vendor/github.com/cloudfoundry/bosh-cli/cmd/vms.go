package cmd

import (
	"fmt"

	"code.cloudfoundry.org/workpool"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type VMsCmd struct {
	ui       boshui.UI
	director boshdir.Director
	parallel int
}

type deploymentInfo struct {
	depName string
	vmInfos []boshdir.VMInfo
}

func NewVMsCmd(ui boshui.UI, director boshdir.Director, parallel int) VMsCmd {
	return VMsCmd{ui: ui, director: director, parallel: parallel}
}

func (c VMsCmd) Run(opts VMsOpts) error {
	instTable := InstanceTable{
		// VMs command should always show VM specifics
		VMDetails: true,

		Details:         false,
		DNS:             opts.DNS,
		Vitals:          opts.Vitals,
		CloudProperties: opts.CloudProperties,
	}

	if len(opts.Deployment) > 0 {
		dep, err := c.director.FindDeployment(opts.Deployment)
		if err != nil {
			return err
		}

		vmInfos, err := dep.VMInfos()
		if err != nil {
			return err
		}

		c.printDeployment(dep, instTable, vmInfos)
		return nil
	}

	return c.printDeployments(instTable, c.parallel)
}

func (c VMsCmd) printDeployments(instTable InstanceTable, parallel int) error {
	deployments, err := c.director.Deployments()
	if err != nil {
		return err
	}

	vmInfos, err := parallelVMInfos(deployments, parallel)

	for _, dep := range deployments {
		if vmInfo, ok := vmInfos[dep.Name()]; ok {
			c.printDeployment(dep, instTable, vmInfo)
		}
	}

	return err
}

func parallelVMInfos(deployments []boshdir.Deployment, parallel int) (map[string][]boshdir.VMInfo, error) {
	if parallel == 0 {
		parallel = 1
	}
	workSize := len(deployments)
	resultc := make(chan deploymentInfo, workSize)
	errorc := make(chan error, workSize)
	defer close(resultc)
	defer close(errorc)
	works := make([]func(), workSize)

	for i, dep := range deployments {
		dep := dep
		works[i] = func() {
			vmInfos, err := dep.VMInfos()
			errorc <- err
			resultc <- deploymentInfo{dep.Name(), vmInfos}
		}
	}

	throttler, err := workpool.NewThrottler(parallel, works)
	if err != nil {
		return nil, err
	}

	throttler.Work()
	vms := make(map[string][]boshdir.VMInfo, workSize)
	var vmInfoErrors []error

	for i := 0; i < workSize; i++ {
		errc := <-errorc
		result := <-resultc
		if errc != nil {
			vmInfoErrors = append(vmInfoErrors, errc)
		} else {
			vms[result.depName] = result.vmInfos
		}
	}

	if len(vmInfoErrors) > 0 {
		err = bosherr.NewMultiError(vmInfoErrors...)
	}

	return vms, err
}

func (c VMsCmd) printDeployment(dep boshdir.Deployment, instTable InstanceTable, vmInfos []boshdir.VMInfo) {
	table := boshtbl.Table{
		Title: fmt.Sprintf("Deployment '%s'", dep.Name()),

		Content: "vms",

		Header: instTable.Headers(),

		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, info := range vmInfos {
		row := instTable.AsValues(instTable.ForVMInfo(info))

		table.Rows = append(table.Rows, row)
	}

	c.ui.PrintTable(table)
}
