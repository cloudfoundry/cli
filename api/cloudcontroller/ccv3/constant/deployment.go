package constant

// DeploymentState describes the states a zero down time deployment used to
// push new apps without restart can be in.
type DeploymentState string

const (
	DeploymentDeploying DeploymentState = "DEPLOYING"
	DeploymentCanceled  DeploymentState = "CANCELED"
	DeploymentDeployed  DeploymentState = "DEPLOYED"
	DeploymentCanceling DeploymentState = "CANCELING"
	DeploymentFailing   DeploymentState = "FAILING"
	DeploymentFailed    DeploymentState = "FAILED"
)

// DeploymentStatusReason describes the status reasons a deployment can have
type DeploymentStatusReason string

const (
	DeploymentStatusReasonDeploying  DeploymentStatusReason = "DEPLOYING"
	DeploymentStatusReasonCanceling  DeploymentStatusReason = "CANCELING"
	DeploymentStatusReasonDeployed   DeploymentStatusReason = "DEPLOYED"
	DeploymentStatusReasonCanceled   DeploymentStatusReason = "CANCELED"
	DeploymentStatusReasonSuperseded DeploymentStatusReason = "SUPERSEDED"
	DeploymentStatusReasonPaused     DeploymentStatusReason = "PAUSED"
)

type DeploymentStatusValue string

const (
	DeploymentStatusValueActive    DeploymentStatusValue = "ACTIVE"
	DeploymentStatusValueFinalized DeploymentStatusValue = "FINALIZED"
)
