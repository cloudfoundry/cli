package api

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type StagingSecurityGroupsRepo interface {
	AddToDefaultStagingSet(group models.ApplicationSecurityGroupFields) error
}

type cloudControllerStagingSecurityGroupRepo struct {
	configRepo configuration.Reader
	gateway    net.Gateway
}

func NewStagingSecurityGroupsRepo(configRepo configuration.Reader, gateway net.Gateway) StagingSecurityGroupsRepo {
	return &cloudControllerStagingSecurityGroupRepo{
		configRepo: configRepo,
		gateway:    gateway,
	}
}

func (repo *cloudControllerStagingSecurityGroupRepo) AddToDefaultStagingSet(group models.ApplicationSecurityGroupFields) error {
	path := fmt.Sprintf("%s/v2/config/staging_security_groups/%s", repo.configRepo.ApiEndpoint(), group.Guid)
	return repo.gateway.CreateResourceFromStruct(path, "")
}
