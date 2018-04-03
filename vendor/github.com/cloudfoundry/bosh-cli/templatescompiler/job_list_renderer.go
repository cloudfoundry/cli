package templatescompiler

import (
	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type JobListRenderer interface {
	Render(
		releaseJobs []bireljob.Job,
		releaseJobProperties map[string]*biproperty.Map,
		jobProperties biproperty.Map,
		globalProperties biproperty.Map,
		deploymentName string,
		address string,
	) (RenderedJobList, error)
}

type jobListRenderer struct {
	jobRenderer JobRenderer
	logger      boshlog.Logger
	logTag      string
}

func NewJobListRenderer(
	jobRenderer JobRenderer,
	logger boshlog.Logger,
) JobListRenderer {
	return &jobListRenderer{
		jobRenderer: jobRenderer,
		logger:      logger,
		logTag:      "jobListRenderer",
	}
}

func (r *jobListRenderer) Render(
	releaseJobs []bireljob.Job,
	releaseJobProperties map[string]*biproperty.Map,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	address string,
) (RenderedJobList, error) {
	r.logger.Debug(r.logTag, "Rendering job list: deploymentName='%s' jobProperties=%#v globalProperties=%#v", deploymentName, jobProperties, globalProperties)
	renderedJobList := NewRenderedJobList()

	// render all the jobs' templates
	for _, releaseJob := range releaseJobs {
		renderedJob, err := r.jobRenderer.Render(releaseJob, releaseJobProperties[releaseJob.Name()], jobProperties, globalProperties, deploymentName, address)
		if err != nil {
			defer renderedJobList.DeleteSilently()
			return renderedJobList, bosherr.WrapErrorf(err, "Rendering templates for job '%s/%s'", releaseJob.Name(), releaseJob.Fingerprint())
		}
		renderedJobList.Add(renderedJob)
	}

	return renderedJobList, nil
}
