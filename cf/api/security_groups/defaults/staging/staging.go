package staging

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"

	. "github.com/cloudfoundry/cli/cf/api/security_groups/defaults"
)

const urlPath = "/v2/config/staging_security_groups"

type StagingSecurityGroupsRepo interface {
	BindToStagingSet(string) error
	List() ([]models.SecurityGroupFields, error)
	UnbindFromStagingSet(string) error
}

type cloudControllerStagingSecurityGroupRepo struct {
	repoBase DefaultSecurityGroupsRepoBase
}

func NewStagingSecurityGroupsRepo(configRepo configuration.Reader, gateway net.Gateway) StagingSecurityGroupsRepo {
	return &cloudControllerStagingSecurityGroupRepo{
		repoBase: DefaultSecurityGroupsRepoBase{
			ConfigRepo: configRepo,
			Gateway:    gateway,
		},
	}
}

func (repo *cloudControllerStagingSecurityGroupRepo) BindToStagingSet(groupGuid string) error {
	return repo.repoBase.Bind(groupGuid, urlPath)
}

func (repo *cloudControllerStagingSecurityGroupRepo) List() ([]models.SecurityGroupFields, error) {
	return repo.repoBase.List(urlPath)
}

func (repo *cloudControllerStagingSecurityGroupRepo) UnbindFromStagingSet(groupGuid string) error {
	return repo.repoBase.Delete(groupGuid, urlPath)
}
