package director

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type VMResp struct {
	VMCID string `json:"vm_cid"`
}

func (d DeploymentImpl) DeleteVM(cid string) error {
	err := d.client.DeleteVM(cid)

	if err != nil {
		resps, listErr := d.client.VMs()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.VMCID == cid {
				return err
			}
		}
	}

	return nil
}

func (c Client) DeleteVM(cid string) error {
	if len(cid) == 0 {
		return bosherr.Error("Expected non-empty VM CID")
	}

	path := fmt.Sprintf("/vms/%s", cid)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(
			err, "Deleting VM '%s'", cid)
	}

	return nil
}

func (c Client) VMs() ([]VMResp, error) {
	var vms []VMResp

	err := c.clientRequest.Get("/vms", &vms)
	if err != nil {
		return vms, bosherr.WrapErrorf(
			err, "Listing VMs")
	}

	return vms, nil
}
