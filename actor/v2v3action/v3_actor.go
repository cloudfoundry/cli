package v2v3action

import (
	"code.cloudfoundry.org/cli/actor/v3action"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . V3Actor

type V3Actor interface {
	ManifestV3Actor
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (v3action.ApplicationSummary, v3action.Warnings, error)
	GetOrganizationByName(orgName string) (v3action.Organization, v3action.Warnings, error)

	CloudControllerAPIVersion() string
}
