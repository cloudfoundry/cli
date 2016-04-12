package spaces

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter . SecurityGroupSpaceBinder

type SecurityGroupSpaceBinder interface {
	BindSpace(securityGroupGuid string, spaceGuid string) error
	UnbindSpace(securityGroupGuid string, spaceGuid string) error
}

type securityGroupSpaceBinder struct {
	configRepo coreconfig.Reader
	gateway    net.Gateway
}

func NewSecurityGroupSpaceBinder(configRepo coreconfig.Reader, gateway net.Gateway) (binder securityGroupSpaceBinder) {
	return securityGroupSpaceBinder{
		configRepo: configRepo,
		gateway:    gateway,
	}
}

func (repo securityGroupSpaceBinder) BindSpace(securityGroupGuid string, spaceGuid string) error {
	url := fmt.Sprintf("/v2/security_groups/%s/spaces/%s",
		securityGroupGuid,
		spaceGuid,
	)

	return repo.gateway.UpdateResourceFromStruct(repo.configRepo.ApiEndpoint(), url, models.SecurityGroupParams{})
}

func (repo securityGroupSpaceBinder) UnbindSpace(securityGroupGuid string, spaceGuid string) error {
	url := fmt.Sprintf("/v2/security_groups/%s/spaces/%s",
		securityGroupGuid,
		spaceGuid,
	)

	return repo.gateway.DeleteResource(repo.configRepo.ApiEndpoint(), url)
}
