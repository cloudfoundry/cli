package staging

import (
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/models"
	"code.cloudfoundry.org/cli/v8/cf/net"

	. "code.cloudfoundry.org/cli/v8/cf/api/securitygroups/defaults"
)

const urlPath = "/v2/config/staging_security_groups"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecurityGroupsRepo

type SecurityGroupsRepo interface {
	BindToStagingSet(string) error
	List() ([]models.SecurityGroupFields, error)
	UnbindFromStagingSet(string) error
}

type cloudControllerStagingSecurityGroupRepo struct {
	repoBase DefaultSecurityGroupsRepoBase
}

func NewSecurityGroupsRepo(configRepo coreconfig.Reader, gateway net.Gateway) SecurityGroupsRepo {
	return &cloudControllerStagingSecurityGroupRepo{
		repoBase: DefaultSecurityGroupsRepoBase{
			ConfigRepo: configRepo,
			Gateway:    gateway,
		},
	}
}

func (repo *cloudControllerStagingSecurityGroupRepo) BindToStagingSet(groupGUID string) error {
	return repo.repoBase.Bind(groupGUID, urlPath)
}

func (repo *cloudControllerStagingSecurityGroupRepo) List() ([]models.SecurityGroupFields, error) {
	return repo.repoBase.List(urlPath)
}

func (repo *cloudControllerStagingSecurityGroupRepo) UnbindFromStagingSet(groupGUID string) error {
	return repo.repoBase.Delete(groupGUID, urlPath)
}
