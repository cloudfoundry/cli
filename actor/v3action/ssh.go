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

	summary, warnings, err := actor.GetApplicationSummaryByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return warnings, err
	}

	var processGUID string
	for _, process := range summary.ProcessSummaries {
		if process.Type == processType {
			processGUID = process.GUID
			if uint(process.Instances.Value) < processIndex+1 {
				return warnings, actionerror.ProcessInstanceNotFoundError{ProcessType: processType, InstanceIndex: processIndex}
			}
			break
		}
	}

	if processGUID == "" {
		return warnings, actionerror.ProcessNotFoundError{ProcessType: processType}
	}

	if !summary.Application.Started() {
		return warnings, actionerror.ApplicationNotStartedError{Name: appName}
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
		Username:              fmt.Sprintf("cf:%s/%d", processGUID, processIndex),
	})

	return warnings, err
}
