package templatescompiler

import (
	"os"
	"path/filepath"

	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bierbrenderer "github.com/cloudfoundry/bosh-cli/templatescompiler/erbrenderer"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type JobRenderer interface {
	Render(releaseJob bireljob.Job, releaseJobProperties *biproperty.Map, jobProperties biproperty.Map, globalProperties biproperty.Map, deploymentName string, address string) (RenderedJob, error)
}

type jobRenderer struct {
	erbRenderer bierbrenderer.ERBRenderer
	fs          boshsys.FileSystem
	uuidGen     boshuuid.Generator
	logger      boshlog.Logger
	logTag      string
}

func NewJobRenderer(
	erbRenderer bierbrenderer.ERBRenderer,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) JobRenderer {
	return &jobRenderer{
		erbRenderer: erbRenderer,
		fs:          fs,
		uuidGen:     uuidGen,
		logger:      logger,
		logTag:      "jobRenderer",
	}
}

func (r *jobRenderer) Render(releaseJob bireljob.Job, releaseJobProperties *biproperty.Map, jobProperties biproperty.Map, globalProperties biproperty.Map, deploymentName string, address string) (RenderedJob, error) {
	context := NewJobEvaluationContext(releaseJob, releaseJobProperties, jobProperties, globalProperties, deploymentName, address, r.uuidGen, r.logger)

	sourcePath := releaseJob.ExtractedPath()

	destinationPath, err := r.fs.TempDir("rendered-jobs")
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating rendered job directory")
	}

	renderedJob := NewRenderedJob(releaseJob, destinationPath, r.fs, r.logger)

	for src, dst := range releaseJob.Templates {
		err := r.renderFile(
			filepath.Join(sourcePath, "templates", src),
			filepath.Join(destinationPath, dst),
			context,
		)
		if err != nil {
			defer renderedJob.DeleteSilently()
			return nil, bosherr.WrapErrorf(err, "Rendering template src: %s, dst: %s", src, dst)
		}
	}

	err = r.renderFile(
		filepath.Join(sourcePath, "monit"),
		filepath.Join(destinationPath, "monit"),
		context,
	)
	if err != nil {
		defer renderedJob.DeleteSilently()
		return nil, bosherr.WrapError(err, "Rendering monit file")
	}

	return renderedJob, nil
}

func (r *jobRenderer) renderFile(sourcePath, destinationPath string, context bierbrenderer.TemplateEvaluationContext) error {
	err := r.fs.MkdirAll(filepath.Dir(destinationPath), os.ModePerm)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating tempdir '%s'", filepath.Dir(destinationPath))
	}

	err = r.erbRenderer.Render(sourcePath, destinationPath, context)
	if err != nil {
		return bosherr.WrapErrorf(err, "Rendering template src: %s, dst: %s", sourcePath, destinationPath)
	}
	return nil
}
