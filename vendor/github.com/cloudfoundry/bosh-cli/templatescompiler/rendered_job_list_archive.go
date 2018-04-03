package templatescompiler

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type RenderedJobListArchive interface {
	List() RenderedJobList
	Path() string
	SHA1() string
	Delete() error
	DeleteSilently()
}

type renderedJobListArchive struct {
	list   RenderedJobList
	path   string
	sha1   string
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

func NewRenderedJobListArchive(
	list RenderedJobList,
	path string,
	sha1 string,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) RenderedJobListArchive {
	return &renderedJobListArchive{
		list:   list,
		path:   path,
		sha1:   sha1,
		fs:     fs,
		logger: logger,
		logTag: "renderedJobListArchive",
	}
}

func (a *renderedJobListArchive) List() RenderedJobList {
	return a.list
}

func (a *renderedJobListArchive) Path() string {
	return a.path
}

func (a *renderedJobListArchive) SHA1() string {
	return a.sha1
}

// Delete removes the archive file (does not delete the rendered jobs in the list)
func (a *renderedJobListArchive) Delete() error {
	err := a.fs.RemoveAll(a.path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting rendered job list archive '%s'", a.path)
	}
	return nil
}

// DeleteSilently removes the archive file (does not delete the rendered jobs in the list),
// logging instead of returning an error.
func (a *renderedJobListArchive) DeleteSilently() {
	err := a.Delete()
	if err != nil {
		a.logger.Error(a.logTag, "Failed to delete rendered job list archive: %s", err.Error())
	}
}

func (a *renderedJobListArchive) String() string {
	return fmt.Sprintf("renderedJobListArchive{list: %s, path: '%s'}", a.list, a.path)
}
