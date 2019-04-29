package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
)

//go:generate counterfeiter . OrgActor

type OrgActor interface {
	GetOrganizationByName(name string) (v7action.Organization, v7action.Warnings, error)
}
