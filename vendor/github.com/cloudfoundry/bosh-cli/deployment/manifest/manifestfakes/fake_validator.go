package manifestfakes

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	biinstall "github.com/cloudfoundry/bosh-cli/installation"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/release/set/manifest"
)

type FakeValidator struct {
	ValidateInputs             []ValidateInput
	ValidateReleaseJobsInputs  []ValidateReleaseJobsInput
	validateOutputs            []ValidateOutput
	validateReleaseJobsOutputs []ValidateReleaseJobsOutput
}

func NewFakeValidator() *FakeValidator {
	return &FakeValidator{
		ValidateInputs:             []ValidateInput{},
		ValidateReleaseJobsInputs:  []ValidateReleaseJobsInput{},
		validateOutputs:            []ValidateOutput{},
		validateReleaseJobsOutputs: []ValidateReleaseJobsOutput{},
	}
}

type ValidateInput struct {
	Manifest           bideplmanifest.Manifest
	ReleaseSetManifest birelsetmanifest.Manifest
}

type ValidateReleaseJobsInput struct {
	Manifest       bideplmanifest.Manifest
	ReleaseManager biinstall.ReleaseManager
}

type ValidateOutput struct {
	Err error
}

type ValidateReleaseJobsOutput struct {
	Err error
}

func (v *FakeValidator) Validate(manifest bideplmanifest.Manifest, releaseSetManifest birelsetmanifest.Manifest) error {
	v.ValidateInputs = append(v.ValidateInputs, ValidateInput{
		Manifest:           manifest,
		ReleaseSetManifest: releaseSetManifest,
	})

	if len(v.validateOutputs) == 0 {
		return bosherr.Errorf("Unexpected FakeValidator.Validate(manifest, releaseSetManifest) called with manifest: %#v", manifest)
	}
	validateOutput := v.validateOutputs[0]
	v.validateOutputs = v.validateOutputs[1:]
	return validateOutput.Err
}

func (v *FakeValidator) ValidateReleaseJobs(manifest bideplmanifest.Manifest, releaseManager biinstall.ReleaseManager) error {
	v.ValidateReleaseJobsInputs = append(v.ValidateReleaseJobsInputs, ValidateReleaseJobsInput{
		Manifest:       manifest,
		ReleaseManager: releaseManager,
	})

	if len(v.validateReleaseJobsOutputs) == 0 {
		return bosherr.Errorf("Unexpected FakeValidator.ValidateReleaseJobs(manifest, releaseManager) called with manifest: %#v", manifest)
	}
	validateReleaseJobsOutput := v.validateReleaseJobsOutputs[0]
	v.validateReleaseJobsOutputs = v.validateReleaseJobsOutputs[1:]
	return validateReleaseJobsOutput.Err
}

func (v *FakeValidator) SetValidateBehavior(outputs []ValidateOutput) {
	v.validateOutputs = outputs
}

func (v *FakeValidator) SetValidateReleaseJobsBehavior(outputs []ValidateReleaseJobsOutput) {
	v.validateReleaseJobsOutputs = outputs
}
