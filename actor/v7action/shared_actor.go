package v7action

import "code.cloudfoundry.org/cli/v8/actor/sharedaction"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SharedActor

type SharedActor interface {
	GatherArchiveResources(archivePath string) ([]sharedaction.Resource, error)
	GatherDirectoryResources(sourceDir string) ([]sharedaction.Resource, error)
	ZipArchiveResources(sourceArchivePath string, filesToInclude []sharedaction.Resource) (string, error)
	ZipDirectoryResources(sourceDir string, filesToInclude []sharedaction.Resource) (string, error)
}
