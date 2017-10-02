package v3action

import "code.cloudfoundry.org/cli/actor/sharedaction"

//go:generate counterfeiter . SharedActor

type SharedActor interface {
	GatherArchiveResources(archivePath string) ([]sharedaction.Resource, error)
	GatherDirectoryResources(sourceDir string) ([]sharedaction.Resource, error)
	ZipArchiveResources(sourceArchivePath string, filesToInclude []sharedaction.Resource) (string, error)
	ZipDirectoryResources(sourceDir string, filesToInclude []sharedaction.Resource) (string, error)
}
