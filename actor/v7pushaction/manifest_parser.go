package v7pushaction

import "code.cloudfoundry.org/cli/util/manifestparser"

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	Apps() []manifestparser.Application
	ContainsManifest() bool
	FullRawManifest() []byte
	RawAppManifest(appName string) ([]byte, error)
}
