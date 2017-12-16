package v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
)

type SSHAuthentication struct {
	Endpoint           string
	HostKeyFingerprint string
	Passcode           string
	Username           string
}

// GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex returns
// back the SSH authentication information for the SSH session.
func (actor Actor) GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex(
	appName string, spaceGUID string, processType string, processIndex uint,
) (SSHAuthentication, Warnings, error) {
	endpoint := actor.CloudControllerClient.AppSSHEndpoint()
	if endpoint == "" {
		return SSHAuthentication{}, nil, actionerror.SSHEndpointNotSetError{}
	}

	fingerprint := actor.CloudControllerClient.AppSSHHostKeyFingerprint()
	if fingerprint == "" {
		return SSHAuthentication{}, nil, actionerror.SSHHostKeyFingerprintNotSetError{}
	}

	passcode, err := actor.UAAClient.GetSSHPasscode(actor.Config.AccessToken(), actor.Config.SSHOAuthClient())
	if err != nil {
		return SSHAuthentication{}, Warnings{}, err
	}

	appSummary, warnings, err := actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return SSHAuthentication{}, warnings, err
	}

	var processSummary ProcessSummary
	for _, appProcessSummary := range appSummary.ProcessSummaries {
		if appProcessSummary.Type == processType {
			processSummary = appProcessSummary
			break
		}
	}
	if processSummary.GUID == "" {
		return SSHAuthentication{}, warnings, actionerror.ProcessNotFoundError{ProcessType: processType}
	}

	if !appSummary.Application.Started() {
		return SSHAuthentication{}, warnings, actionerror.ApplicationNotStartedError{Name: appName}
	}

	var processInstance ProcessInstance
	for _, instance := range processSummary.InstanceDetails {
		if uint(instance.Index) == processIndex {
			processInstance = instance
			break
		}
	}

	if processInstance == (ProcessInstance{}) {
		return SSHAuthentication{}, warnings, actionerror.ProcessInstanceNotFoundError{ProcessType: processType, InstanceIndex: processIndex}
	}

	if !processInstance.Running() {
		return SSHAuthentication{}, warnings, actionerror.ProcessInstanceNotRunningError{ProcessType: processType,
			InstanceIndex: processIndex}
	}

	return SSHAuthentication{
		Endpoint:           endpoint,
		HostKeyFingerprint: fingerprint,
		Passcode:           passcode,
		Username:           fmt.Sprintf("cf:%s/%d", processSummary.GUID, processIndex),
	}, warnings, err
}
