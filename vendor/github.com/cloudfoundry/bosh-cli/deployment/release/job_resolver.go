package release

import (
	birel "github.com/cloudfoundry/bosh-cli/release"
	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type JobResolver interface {
	Resolve(jobName, releaseName string) (bireljob.Job, error)
}

type resolver struct {
	releaseManager birel.Manager
}

func NewJobResolver(releaseManager birel.Manager) JobResolver {
	return &resolver{
		releaseManager: releaseManager,
	}
}

func (r *resolver) Resolve(jobName, releaseName string) (bireljob.Job, error) {
	release, found := r.releaseManager.Find(releaseName)
	if !found {
		return bireljob.Job{}, bosherr.Errorf("Finding release '%s'", releaseName)
	}

	releaseJob, found := release.FindJobByName(jobName)
	if !found {
		return bireljob.Job{}, bosherr.Errorf("Finding job '%s' in release '%s'", jobName, releaseName)
	}

	return releaseJob, nil
}
