package spaces

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type SecurityGroupSpaceBinder interface {
	BindSpace(securityGroupGuid string, spaceGuid string) error
	UnbindSpace(securityGroupGuid string, spaceGuid string) error
}

type securityGroupSpaceBinder struct {
	configRepo core_config.Reader
	gateway    net.Gateway
}

func NewSecurityGroupSpaceBinder(configRepo core_config.Reader, gateway net.Gateway) (binder securityGroupSpaceBinder) {
	return securityGroupSpaceBinder{
		configRepo: configRepo,
		gateway:    gateway,
	}
}

func (repo securityGroupSpaceBinder) BindSpace(securityGroupGuid string, spaceGuid string) error {
	url := fmt.Sprintf("%s/v2/security_groups/%s/spaces/%s",
		repo.configRepo.ApiEndpoint(),
		securityGroupGuid,
		spaceGuid,
	)

	return repo.gateway.UpdateResourceFromStruct(url, models.SecurityGroupParams{})
}

func (repo securityGroupSpaceBinder) UnbindSpace(securityGroupGuid string, spaceGuid string) error {
	url := fmt.Sprintf("%s/v2/security_groups/%s/spaces/%s",
		repo.configRepo.ApiEndpoint(),
		securityGroupGuid,
		spaceGuid,
	)

	return repo.gateway.DeleteResource(url)
}
