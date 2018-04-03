package cmd

import (
	"fmt"

	"code.cloudfoundry.org/workpool"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type InstancesCmd struct {
	ui       boshui.UI
	director boshdir.Director
	parallel int
}

func NewInstancesCmd(ui boshui.UI, director boshdir.Director, parallel int) InstancesCmd {
	return InstancesCmd{ui: ui, director: director, parallel: parallel}
}

func (c InstancesCmd) Run(opts InstancesOpts) error {
	instTable := InstanceTable{
		Processes: opts.Processes,
		Details:   opts.Details,
		DNS:       opts.DNS,
		Vitals:    opts.Vitals,
	}

	if len(opts.Deployment) > 0 {
		dep, err := c.director.FindDeployment(opts.Deployment)
		if err != nil {
			return err
		}

		instanceInfos, err := dep.InstanceInfos()
		if err != nil {
			return err
		}

		c.printDeployment(dep, instTable, opts, instanceInfos)
		return nil
	}

	return c.printDeployments(instTable, opts)
}

func (c InstancesCmd) printDeployments(instTable InstanceTable, opts InstancesOpts) error {
	deployments, err := c.director.Deployments()
	if err != nil {
		return err
	}

	instanceInfos, err := parallelInstanceInfos(deployments, c.parallel)

	for _, dep := range deployments {
		if instanceInfo, ok := instanceInfos[dep.Name()]; ok {
			c.printDeployment(dep, instTable, opts, instanceInfo)
		}
	}

	return err
}

func parallelInstanceInfos(deployments []boshdir.Deployment, parallel int) (map[string][]boshdir.VMInfo, error) {
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
			instanceInfos, err := dep.InstanceInfos()
			errorc <- err
			resultc <- deploymentInfo{dep.Name(), instanceInfos}
		}
	}

	throttler, err := workpool.NewThrottler(parallel, works)
	if err != nil {
		return nil, err
	}
	throttler.Work()
	vms := make(map[string][]boshdir.VMInfo, workSize)
	var instanceInfoErrors []error
	for i := 0; i < workSize; i++ {
		errc := <-errorc
		result := <-resultc
		if errc != nil {
			instanceInfoErrors = append(instanceInfoErrors, errc)
		} else {
			vms[result.depName] = result.vmInfos
		}
	}
	if len(instanceInfoErrors) > 0 {
		err = bosherr.NewMultiError(instanceInfoErrors...)
	}
	return vms, err
}

func (c InstancesCmd) printDeployment(dep boshdir.Deployment, instTable InstanceTable, opts InstancesOpts, instanceInfos []boshdir.VMInfo) {
	table := boshtbl.Table{
		Title: fmt.Sprintf("Deployment '%s'", dep.Name()),

		Content: "instances",

		Header: instTable.Headers(),

		SortBy: []boshtbl.ColumnSort{
			{Column: 0, Asc: true},
			{Column: 1, Asc: true}, // sort by process so that VM row is first
		},
	}

	for _, info := range instanceInfos {
		if opts.Failing && info.IsRunning() {
			continue
		}

		row := instTable.AsValues(instTable.ForVMInfo(info))

		section := boshtbl.Section{
			FirstColumn: row[0],
			Rows:        [][]boshtbl.Value{row},
		}

		if opts.Processes {
			for _, p := range info.Processes {
				if opts.Failing && p.IsRunning() {
					continue
				}

				row := instTable.AsValues(instTable.ForProcess(p))

				section.Rows = append(section.Rows, row)
			}
		}

		table.Sections = append(table.Sections, section)
	}

	c.ui.PrintTable(table)
}
