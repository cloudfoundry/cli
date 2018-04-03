package templatescompiler

import (
	"fmt"

	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type RenderedJob interface {
	Job() bireljob.Job
	Path() string // dir of multiple rendered files
	Delete() error
	DeleteSilently()
}

type renderedJob struct {
	job    bireljob.Job
	path   string
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

func NewRenderedJob(
	job bireljob.Job,
	path string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) RenderedJob {
	return &renderedJob{
		job:    job,
		path:   path,
		fs:     fs,
		logger: logger,
		logTag: "renderedJob",
	}
}

func (j *renderedJob) Job() bireljob.Job { return j.job }

// Path returns a parent directory with one or more sub-dirs for each job, each with one or more rendered template files
func (j *renderedJob) Path() string { return j.path }

func (j *renderedJob) Delete() error {
	err := j.fs.RemoveAll(j.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting rendered job '%s' tarball '%s'", j.job.Name, j.path)
	}
	return nil
}

func (j *renderedJob) DeleteSilently() {
	err := j.Delete()
	if err != nil {
		j.logger.Error(j.logTag, "Failed to delete rendered job: %s", err.Error())
	}
}

func (j *renderedJob) String() string {
	return fmt.Sprintf("renderedJob{job: '%s', path: '%s'}", j.job.Name(), j.path)
}
