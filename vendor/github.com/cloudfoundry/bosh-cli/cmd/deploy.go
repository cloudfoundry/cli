package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type DeployCmd struct {
	ui              boshui.UI
	deployment      boshdir.Deployment
	releaseUploader ReleaseUploader
}

type ReleaseUploader interface {
	UploadReleases([]byte) ([]byte, error)
}

func NewDeployCmd(
	ui boshui.UI,
	deployment boshdir.Deployment,
	releaseUploader ReleaseUploader,
) DeployCmd {
	return DeployCmd{ui, deployment, releaseUploader}
}

func (c DeployCmd) Run(opts DeployOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating manifest")
	}

	err = c.checkDeploymentName(bytes)
	if err != nil {
		return err
	}

	bytes, err = c.releaseUploader.UploadReleases(bytes)
	if err != nil {
		return err
	}

	deploymentDiff, err := c.deployment.Diff(bytes, opts.NoRedact)
	if err != nil {
		return err
	}

	diff := NewDiff(deploymentDiff.Diff)
	diff.Print(c.ui)

	err = c.ui.AskForConfirmation()
	if err != nil {
		return err
	}

	updateOpts := boshdir.UpdateOpts{
		Recreate:    opts.Recreate,
		Fix:         opts.Fix,
		SkipDrain:   opts.SkipDrain,
		DryRun:      opts.DryRun,
		Canaries:    opts.Canaries,
		MaxInFlight: opts.MaxInFlight,
		Diff:        deploymentDiff,
	}

	return c.deployment.Update(bytes, updateOpts)
}

func (c DeployCmd) checkDeploymentName(bytes []byte) error {
	manifest, err := boshdir.NewManifestFromBytes(bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Parsing manifest")
	}

	if manifest.Name != c.deployment.Name() {
		errMsg := "Expected manifest to specify deployment name '%s' but was '%s'"
		return bosherr.Errorf(errMsg, c.deployment.Name(), manifest.Name)
	}

	return nil
}
