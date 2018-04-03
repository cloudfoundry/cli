package templatescompiler

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type RenderedJobList interface {
	Add(RenderedJob)
	All() []RenderedJob
	Delete() error
	DeleteSilently()
}

type renderedJobList struct {
	renderedJobs []RenderedJob
	logTag       string
}

func NewRenderedJobList() RenderedJobList {
	return &renderedJobList{
		renderedJobs: []RenderedJob{},
	}
}

func (j *renderedJobList) Add(renderedJob RenderedJob) {
	j.renderedJobs = append(j.renderedJobs, renderedJob)
}

func (j *renderedJobList) All() []RenderedJob {
	return append([]RenderedJob{}, j.renderedJobs...)
}

func (j *renderedJobList) Delete() error {
	for _, renderedJob := range j.renderedJobs {
		err := renderedJob.Delete()
		if err != nil {
			return bosherr.WrapErrorf(err, "Deleting rendered job '%s'", renderedJob.Job().Name())
		}
	}

	j.renderedJobs = []RenderedJob{}

	return nil
}

func (j *renderedJobList) DeleteSilently() {
	for _, renderedJob := range j.renderedJobs {
		renderedJob.DeleteSilently()
	}
	j.renderedJobs = []RenderedJob{}
}

func (j *renderedJobList) String() string {
	return fmt.Sprintf("renderedJobList{renderedJobs: '%s'}", j.renderedJobs)
}
