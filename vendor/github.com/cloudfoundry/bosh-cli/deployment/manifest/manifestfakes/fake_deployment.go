package manifestfakes

import (
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

func NewFakeReleaseJobRef() bideplmanifest.ReleaseJobRef {
	return bideplmanifest.ReleaseJobRef{
		Name:    "fake-release-job-ref-name",
		Release: "fake-release-job-ref-release",
	}
}

func NewFakeJob() bideplmanifest.Job {
	return bideplmanifest.Job{
		Name:      "fake-deployment-job",
		Instances: 1,
		Templates: []bideplmanifest.ReleaseJobRef{NewFakeReleaseJobRef()},
	}
}

func NewFakeDeployment() bideplmanifest.Manifest {
	return bideplmanifest.Manifest{
		Name:       "fake-deployment-name",
		Properties: biproperty.Map{},
		Jobs:       []bideplmanifest.Job{NewFakeJob()},
	}
}
