package staging

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"

	. "github.com/cloudfoundry/cli/cf/api/security_groups/defaults"
)

type StagingSecurityGroupsRepo interface {
	AddToDefaultStagingSet(string) error
	List() ([]models.SecurityGroupFields, error)
	RemoveFromDefaultStagingSet(string) error
}

type cloudControllerStagingSecurityGroupRepo struct {
	repoBase SecurityGroupsRepoBase
}

func NewStagingSecurityGroupsRepo(configRepo configuration.Reader, gateway net.Gateway) StagingSecurityGroupsRepo {
	return &cloudControllerStagingSecurityGroupRepo{
		repoBase: SecurityGroupsRepoBase{
			ConfigRepo: configRepo,
			Gateway:    gateway,
		},
	}
}

func (repo *cloudControllerStagingSecurityGroupRepo) AddToDefaultStagingSet(groupGuid string) error {
	return repo.repoBase.Add(groupGuid, "/v2/config/staging_security_groups/")
}

func (repo *cloudControllerStagingSecurityGroupRepo) List() ([]models.SecurityGroupFields, error) {
	return repo.repoBase.List("/v2/config/staging_security_groups")
}

func (repo *cloudControllerStagingSecurityGroupRepo) RemoveFromDefaultStagingSet(groupGuid string) error {
	path := fmt.Sprintf("%s/v2/config/staging_security_groups/%s", repo.repoBase.ConfigRepo.ApiEndpoint(), groupGuid)
	return repo.repoBase.Gateway.DeleteResource(path)
}
