package running

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"

	. "github.com/cloudfoundry/cli/cf/api/security_groups/defaults"
)

const urlPath = "/v2/config/running_security_groups"

type RunningSecurityGroupsRepo interface {
	AddToDefaultRunningSet(string) error
}

type cloudControllerRunningSecurityGroupRepo struct {
	repoBase SecurityGroupsRepoBase
}

func NewRunningSecurityGroupsRepo(configRepo configuration.Reader, gateway net.Gateway) RunningSecurityGroupsRepo {
	return &cloudControllerRunningSecurityGroupRepo{
		repoBase: SecurityGroupsRepoBase{
			ConfigRepo: configRepo,
			Gateway:    gateway,
		},
	}
}

func (repo *cloudControllerRunningSecurityGroupRepo) AddToDefaultRunningSet(groupGuid string) error {
	return repo.repoBase.Add(groupGuid, urlPath)
}
