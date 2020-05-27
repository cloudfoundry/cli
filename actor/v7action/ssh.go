package v7action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/resources"
)

type SSHAuthentication struct {
	Endpoint           string
	HostKeyFingerprint string
	Passcode           string
	Username           string
}

func (actor Actor) GetSSHPasscode() (string, error) {
	return actor.UAAClient.GetSSHPasscode(actor.Config.AccessToken(), actor.Config.SSHOAuthClient())
}

// GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex returns
// back the SSH authentication information for the SSH session.
func (actor Actor) GetSecureShellConfigurationByApplicationNameSpaceProcessTypeAndIndex(
	appName string, spaceGUID string, processType string, processIndex uint,
) (SSHAuthentication, Warnings, error) {
	var allWarnings Warnings

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

	application, appWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	allWarnings = append(allWarnings, appWarnings...)
	if err != nil {
		return SSHAuthentication{}, allWarnings, err
	}

	if !application.Started() {
		return SSHAuthentication{}, allWarnings, actionerror.ApplicationNotStartedError{Name: appName}
	}

	username, processWarnings, err := actor.getUsername(application, processType, processIndex)
	allWarnings = append(allWarnings, processWarnings...)
	if err != nil {
		return SSHAuthentication{}, allWarnings, err
	}

	return SSHAuthentication{
		Endpoint:           endpoint,
		HostKeyFingerprint: fingerprint,
		Passcode:           passcode,
		Username:           username,
	}, allWarnings, err
}

func (actor Actor) getUsername(application resources.Application, processType string, processIndex uint) (string, Warnings, error) {
	processSummaries, processWarnings, err := actor.getProcessSummariesForApp(application.GUID, false)
	if err != nil {
		return "", processWarnings, err
	}

	var processSummary ProcessSummary
	for _, appProcessSummary := range processSummaries {
		if appProcessSummary.Type == processType {
			processSummary = appProcessSummary
			break
		}
	}

	if processSummary.GUID == "" {
		return "", processWarnings, actionerror.ProcessNotFoundError{ProcessType: processType}
	}

	var processInstance ProcessInstance
	for _, instance := range processSummary.InstanceDetails {
		if uint(instance.Index) == processIndex {
			processInstance = instance
			break
		}
	}

	if processInstance == (ProcessInstance{}) {
		return "", processWarnings, actionerror.ProcessInstanceNotFoundError{ProcessType: processType, InstanceIndex: processIndex}
	}

	if !processInstance.Running() {
		return "", processWarnings, actionerror.ProcessInstanceNotRunningError{ProcessType: processType, InstanceIndex: processIndex}
	}

	return fmt.Sprintf("cf:%s/%d", processSummary.GUID, processIndex), processWarnings, nil
}
