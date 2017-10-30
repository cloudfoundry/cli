package v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
)

type SSHOptions struct {
	Commands              []string
	Forward               []string
	LocalPortForwardSpecs []sharedaction.LocalPortForward
	SkipHostValidation    bool
	SkipRemoteExecution   bool
	TTYOption             sharedaction.TTYOption
}

func (actor Actor) ExecuteSecureShellByApplicationNameSpaceProcessTypeAndIndex(appName string, spaceGUID string, processType string, processIndex uint, sshOptions SSHOptions) (Warnings, error) {
	endpoint := actor.CloudControllerClient.AppSSHEndpoint()
	if endpoint == "" {
		return nil, actionerror.SSHEndpointNotSetError{}
	}

	fingerprint := actor.CloudControllerClient.AppSSHHostKeyFingerprint()
	if fingerprint == "" {
		return nil, actionerror.SSHHostKeyFingerprintNotSetError{}
	}

	passcode, err := actor.UAAClient.GetSSHPasscode(actor.Config.AccessToken(), actor.Config.SSHOAuthClient())
	if err != nil {
		return Warnings{}, err
	}

	appSummary, warnings, err := actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	var processSummary ProcessSummary
	for _, pS := range appSummary.ProcessSummaries {
		if pS.Type == processType {
			processSummary = pS
			break
		}
	}
	if processSummary.GUID == "" {
		return warnings, actionerror.ProcessNotFoundError{ProcessType: processType}
	}

	if !appSummary.Application.Started() {
		return warnings, actionerror.ApplicationNotStartedError{Name: appName}
	}

	var processInstance Instance
	for _, instance := range processSummary.InstanceDetails {
		if uint(instance.Index) == processIndex {
			processInstance = instance
			break
		}
	}

	if processInstance == (Instance{}) {
		return warnings, actionerror.ProcessInstanceNotFoundError{ProcessType: processType, InstanceIndex: processIndex}
	}

	if !processInstance.Running() {
		return warnings, actionerror.ProcessInstanceNotRunningError{ProcessType: processType,
			InstanceIndex: processIndex}
	}

	err = actor.SharedActor.ExecuteSecureShell(sharedaction.SSHOptions{
		Commands:              sshOptions.Commands,
		Endpoint:              endpoint,
		HostKeyFingerprint:    fingerprint,
		LocalPortForwardSpecs: sshOptions.LocalPortForwardSpecs,
		Passcode:              passcode,
		SkipHostValidation:    sshOptions.SkipHostValidation,
		SkipRemoteExecution:   sshOptions.SkipRemoteExecution,
		TTYOption:             sshOptions.TTYOption,
		Username:              fmt.Sprintf("cf:%s/%d", processSummary.GUID, processIndex),
	})

	return warnings, err
}
