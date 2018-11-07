package v7pushaction

import (
	"io"

	"code.cloudfoundry.org/cli/actor/sharedaction"
)

//go:generate counterfeiter io.ReadCloser
//go:generate counterfeiter . SharedActor

type SharedActor interface {
	GatherArchiveResources(archivePath string) ([]sharedaction.Resource, error)
	GatherDirectoryResources(sourceDir string) ([]sharedaction.Resource, error)
	ReadArchive(archivePath string) (io.ReadCloser, int64, error)
	ZipArchiveResources(sourceArchivePath string, filesToInclude []sharedaction.Resource) (string, error)
	ZipDirectoryResources(sourceDir string, filesToInclude []sharedaction.Resource) (string, error)
}
