package spaces

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/models"
	"code.cloudfoundry.org/cli/v8/cf/net"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecurityGroupSpaceBinder

type SecurityGroupSpaceBinder interface {
	BindSpace(securityGroupGUID string, spaceGUID string) error
	UnbindSpace(securityGroupGUID string, spaceGUID string) error
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

func (repo securityGroupSpaceBinder) BindSpace(securityGroupGUID string, spaceGUID string) error {
	url := fmt.Sprintf("/v2/security_groups/%s/spaces/%s",
		securityGroupGUID,
		spaceGUID,
	)

	return repo.gateway.UpdateResourceFromStruct(repo.configRepo.APIEndpoint(), url, models.SecurityGroupParams{})
}

func (repo securityGroupSpaceBinder) UnbindSpace(securityGroupGUID string, spaceGUID string) error {
	url := fmt.Sprintf("/v2/security_groups/%s/spaces/%s",
		securityGroupGUID,
		spaceGUID,
	)

	return repo.gateway.DeleteResource(repo.configRepo.APIEndpoint(), url)
}
