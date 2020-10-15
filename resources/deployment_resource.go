package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Deployment struct {
	GUID          string
	State         constant.DeploymentState
	StatusValue   constant.DeploymentStatusValue
	StatusReason  constant.DeploymentStatusReason
	RevisionGUID  string
	DropletGUID   string
	CreatedAt     string
	UpdatedAt     string
	Relationships Relationships
	NewProcesses  []Process
}

// MarshalJSON converts a Deployment into a Cloud Controller Deployment.
func (d Deployment) MarshalJSON() ([]byte, error) {
	type Revision struct {
		GUID string `json:"guid,omitempty"`
	}
	type Droplet struct {
		GUID string `json:"guid,omitempty"`
	}

	var ccDeployment struct {
		Droplet       *Droplet      `json:"droplet,omitempty"`
		Revision      *Revision     `json:"revision,omitempty"`
		Relationships Relationships `json:"relationships,omitempty"`
	}

	if d.DropletGUID != "" {
		ccDeployment.Droplet = &Droplet{d.DropletGUID}
	}

	if d.RevisionGUID != "" {
		ccDeployment.Revision = &Revision{d.RevisionGUID}
	}

	ccDeployment.Relationships = d.Relationships

	return json.Marshal(ccDeployment)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Deployment response.
func (d *Deployment) UnmarshalJSON(data []byte) error {
	var ccDeployment struct {
		GUID          string                   `json:"guid,omitempty"`
		CreatedAt     string                   `json:"created_at,omitempty"`
		Relationships Relationships            `json:"relationships,omitempty"`
		State         constant.DeploymentState `json:"state,omitempty"`
		Status        struct {
			Value  constant.DeploymentStatusValue  `json:"value"`
			Reason constant.DeploymentStatusReason `json:"reason"`
		} `json:"status"`
		Droplet      Droplet   `json:"droplet,omitempty"`
		NewProcesses []Process `json:"new_processes,omitempty"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccDeployment)
	if err != nil {
		return err
	}

	d.GUID = ccDeployment.GUID
	d.CreatedAt = ccDeployment.CreatedAt
	d.Relationships = ccDeployment.Relationships
	d.State = ccDeployment.State
	d.StatusValue = ccDeployment.Status.Value
	d.StatusReason = ccDeployment.Status.Reason
	d.DropletGUID = ccDeployment.Droplet.GUID
	d.NewProcesses = ccDeployment.NewProcesses

	return nil
}
