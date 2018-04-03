package director

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Instance struct {
	AgentID string `json:"agent_id"`
	VMID    string `json:"cid"`

	ID    string `json:"id"`
	Group string `json:"job"`

	AZ        string `json:"az"`
	ExpectsVM bool   `json:"expects_vm"`

	IPs []string `json:"ips"`
}

func (d DeploymentImpl) InstanceInfos() ([]VMInfo, error) {
	infos, err := d.client.DeploymentInstanceInfos(d.name)
	if err != nil {
		return nil, err
	}
	return infos, nil
}

func (d DeploymentImpl) Instances() ([]Instance, error) {
	insts, err := d.client.DeploymentInstances(d.name)
	if err != nil {
		return nil, err
	}

	return insts, nil
}

func (c Client) DeploymentInstanceInfos(deploymentName string) ([]VMInfo, error) {
	return c.deploymentResourceInfos(deploymentName, "instances")
}

func (c Client) DeploymentInstances(deploymentName string) ([]Instance, error) {
	if len(deploymentName) == 0 {
		return nil, bosherr.Error("Expected non-empty deployment name")
	}

	var insts []Instance

	path := fmt.Sprintf("/deployments/%s/instances", deploymentName)

	err := c.clientRequest.Get(path, &insts)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding instances")
	}

	return insts, nil
}
