package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Deployment struct {
	GUID             string
	State            constant.DeploymentState
	StatusValue      constant.DeploymentStatusValue
	StatusReason     constant.DeploymentStatusReason
	CanaryStatus     CanaryStatus
	LastStatusChange string
	Options          DeploymentOpts
	RevisionGUID     string
	DropletGUID      string
	CreatedAt        string
	UpdatedAt        string
	Relationships    Relationships
	NewProcesses     []Process
	Strategy         constant.DeploymentStrategy
}

type DeploymentOpts struct {
	MaxInFlight             int                      `json:"max_in_flight,omitempty"`
	CanaryDeploymentOptions *CanaryDeploymentOptions `json:"canary,omitempty"`
}

func (d DeploymentOpts) IsEmpty() bool {
	return d.MaxInFlight == 0 && (d.CanaryDeploymentOptions == nil || len(d.CanaryDeploymentOptions.Steps) == 0)
}

type CanaryDeploymentOptions struct {
	Steps []CanaryStep `json:"steps"`
}

type CanaryStep struct {
	InstanceWeight int64 `json:"instance_weight"`
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
		Droplet       *Droplet                    `json:"droplet,omitempty"`
		Options       *DeploymentOpts             `json:"options,omitempty"`
		Revision      *Revision                   `json:"revision,omitempty"`
		Strategy      constant.DeploymentStrategy `json:"strategy,omitempty"`
		Relationships Relationships               `json:"relationships,omitempty"`
	}

	if d.DropletGUID != "" {
		ccDeployment.Droplet = &Droplet{d.DropletGUID}
	}

	if d.RevisionGUID != "" {
		ccDeployment.Revision = &Revision{d.RevisionGUID}
	}

	if d.Strategy != "" {
		ccDeployment.Strategy = d.Strategy
	}

	if !d.Options.IsEmpty() {
		ccDeployment.Options = &d.Options
	}

	ccDeployment.Relationships = d.Relationships

	return json.Marshal(ccDeployment)
}

type CanaryStepStatus struct {
	CurrentStep int `json:"current"`
	TotalSteps  int `json:"total"`
}

type CanaryStatus struct {
	Steps CanaryStepStatus `json:"steps"`
}

// UnmarshalJSON helps unmarshal a Cloud Controller Deployment response.
func (d *Deployment) UnmarshalJSON(data []byte) error {
	var ccDeployment struct {
		GUID          string                   `json:"guid,omitempty"`
		CreatedAt     string                   `json:"created_at,omitempty"`
		Relationships Relationships            `json:"relationships,omitempty"`
		State         constant.DeploymentState `json:"state,omitempty"`
		Status        struct {
			Details struct {
				LastStatusChange string `json:"last_status_change"`
			}
			Value        constant.DeploymentStatusValue  `json:"value"`
			Reason       constant.DeploymentStatusReason `json:"reason"`
			CanaryStatus CanaryStatus                    `json:"canary,omitempty"`
		} `json:"status"`
		Droplet      Droplet                     `json:"droplet,omitempty"`
		NewProcesses []Process                   `json:"new_processes,omitempty"`
		Strategy     constant.DeploymentStrategy `json:"strategy"`
		Options      DeploymentOpts              `json:"options,omitempty"`
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
	d.CanaryStatus = ccDeployment.Status.CanaryStatus
	d.LastStatusChange = ccDeployment.Status.Details.LastStatusChange
	d.DropletGUID = ccDeployment.Droplet.GUID
	d.NewProcesses = ccDeployment.NewProcesses
	d.Strategy = ccDeployment.Strategy
	d.Options = ccDeployment.Options

	return nil
}
