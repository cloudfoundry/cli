package v7pushaction

import "code.cloudfoundry.org/cli/util/manifestparser"

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	Apps(appName string) ([]manifestparser.Application, error)
	FullRawManifest() []byte
	RawAppManifest(appName string) ([]byte, error)
}
