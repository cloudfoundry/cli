package v7pushaction

import "code.cloudfoundry.org/cli/util/manifestparser"

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	Apps(appName string) ([]manifestparser.Application, error)
	ContainsManifest() bool
	FullRawManifest() []byte
	RawAppManifest(appName string) ([]byte, error)
	ApplyNoRouteOverride(appName string, noRoute bool) error
}
