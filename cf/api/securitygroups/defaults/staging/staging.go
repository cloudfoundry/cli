package staging

import (
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"

	. "github.com/cloudfoundry/cli/cf/api/securitygroups/defaults"
)

const urlPath = "/v2/config/staging_security_groups"

//go:generate counterfeiter . StagingSecurityGroupsRepo

type StagingSecurityGroupsRepo interface {
	BindToStagingSet(string) error
	List() ([]models.SecurityGroupFields, error)
	UnbindFromStagingSet(string) error
}

type cloudControllerStagingSecurityGroupRepo struct {
	repoBase DefaultSecurityGroupsRepoBase
}

func NewStagingSecurityGroupsRepo(configRepo coreconfig.Reader, gateway net.Gateway) StagingSecurityGroupsRepo {
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
