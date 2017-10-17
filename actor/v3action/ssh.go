package v3action

import (
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
)

type SSHOptions struct {
	Commands            []string
	SkipHostValidation  bool
	DisablePseudoTTY    bool
	ForcePseudoTTY      bool
	RequestPseudoTTY    bool
	SkipRemoteExecution bool
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
				return warnings, actionerror.ProcessInstanceNotFoundError{}
			}
			break
		}
	}

	if processGUID == "" {
		return warnings, actionerror.ProcessTypeNotFoundError{Name: processType}
	}

	if !summary.Application.Started() {
		return warnings, actionerror.ApplicationNotStartedError{Name: appName}
	}

	err = actor.SharedActor.ExecuteSecureShell(sharedaction.SSHOptions{
		Username:           fmt.Sprintf("cf:%s/%d", processGUID, processIndex),
		Passcode:           passcode,
		Endpoint:           endpoint,
		HostKeyFingerprint: fingerprint,
	})

	// call sharedactor.sshsomething stuff
	return warnings, err
}
