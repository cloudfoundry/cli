package fakes

type FakeDeploymentRepo struct {
	UpdateCurrentManifestSHA string
	UpdateCurrentErr         error

	findCurrentOutput deploymentRepoFindCurrentOutput
}

type deploymentRepoFindCurrentOutput struct {
	manifestSHA string
	found       bool
	err         error
}

func NewFakeDeploymentRepo() *FakeDeploymentRepo {
	return &FakeDeploymentRepo{}
}

func (r *FakeDeploymentRepo) UpdateCurrent(manifestSHA string) error {
	r.UpdateCurrentManifestSHA = manifestSHA
	return r.UpdateCurrentErr
}

func (r *FakeDeploymentRepo) FindCurrent() (manifestSHA string, found bool, err error) {
	return r.findCurrentOutput.manifestSHA, r.findCurrentOutput.found, r.findCurrentOutput.err
}

func (r *FakeDeploymentRepo) SetFindCurrentBehavior(manifestSHA string, found bool, err error) {
	r.findCurrentOutput = deploymentRepoFindCurrentOutput{
		manifestSHA: manifestSHA,
		found:       found,
		err:         err,
	}
}
