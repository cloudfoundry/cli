package installation

import (
	bideplrel "github.com/cloudfoundry/bosh-cli/deployment/release"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/installation/manifest"
	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type JobResolver interface {
	From(biinstallmanifest.Manifest) ([]bireljob.Job, error)
}

type jobResolver struct {
	releaseJobResolver bideplrel.JobResolver
}

func NewJobResolver(
	releaseJobResolver bideplrel.JobResolver,
) JobResolver {
	return &jobResolver{
		releaseJobResolver: releaseJobResolver,
	}
}

func (b *jobResolver) From(installationManifest biinstallmanifest.Manifest) ([]bireljob.Job, error) {
	// installation only ever has one job: the cpi job.
	jobsReferencesInRelease := []biinstallmanifest.ReleaseJobRef{installationManifest.Template}

	releaseJobs := make([]bireljob.Job, len(jobsReferencesInRelease), len(jobsReferencesInRelease))
	for i, jobRef := range jobsReferencesInRelease {
		release, err := b.releaseJobResolver.Resolve(jobRef.Name, jobRef.Release)
		if err != nil {
			return releaseJobs, bosherr.WrapErrorf(err, "Resolving job '%s' in release '%s'", jobRef.Name, jobRef.Release)
		}
		releaseJobs[i] = release
	}
	return releaseJobs, nil
}
