package api

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/net"
)

type StagingSecurityGroupsRepo interface {
	AddToDefaultStagingSet(string) error
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

func (repo *cloudControllerStagingSecurityGroupRepo) AddToDefaultStagingSet(groupGuid string) error {
	path := fmt.Sprintf("%s/v2/config/staging_security_groups/%s", repo.configRepo.ApiEndpoint(), groupGuid)
	return repo.gateway.CreateResourceFromStruct(path, "")
}
