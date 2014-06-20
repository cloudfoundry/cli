package api

import (
	"fmt"
	"net/url"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type AppSecurityGroup interface {
	Create(models.ApplicationSecurityGroupFields) error
	Delete(string) error
	Read(string) (models.ApplicationSecurityGroupFields, error)
}

type ApplicationSecurityGroupRepo struct {
	gateway net.Gateway
	config  configuration.Reader
}

func NewApplicationSecurityGroupRepo(config configuration.Reader, gateway net.Gateway) ApplicationSecurityGroupRepo {
	return ApplicationSecurityGroupRepo{
		config:  config,
		gateway: gateway,
	}
}

func (repo ApplicationSecurityGroupRepo) Create(groupFields models.ApplicationSecurityGroupFields) error {
	path := fmt.Sprintf("%s/v2/app_security_groups", repo.config.ApiEndpoint())
	return repo.gateway.CreateResourceFromStruct(path, groupFields)
}

func (repo ApplicationSecurityGroupRepo) Read(name string) (models.ApplicationSecurityGroupFields, error) {
	path := fmt.Sprintf("/v2/app_security_groups?q=%s&inline-relations-depth=1", url.QueryEscape("name:"+name))
	group := models.ApplicationSecurityGroupFields{}
	foundGroup := false

	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.ApplicationSecurityGroupResource{},
		func(resource interface{}) bool {
			if g, ok := resource.(resources.ApplicationSecurityGroupResource); ok {
				group = g.ToFields()
				foundGroup = true
			}

			return false
		},
	)
	if err != nil {
		return group, err
	}

	if !foundGroup {
		err = errors.NewModelNotFoundError("application security group", name)
	}

	return group, err
}

func (repo ApplicationSecurityGroupRepo) Delete(securityGroupGuid string) error {
	path := fmt.Sprintf("%s/v2/app_security_groups/%s", repo.config.ApiEndpoint(), securityGroupGuid)
	return repo.gateway.DeleteResource(path)
}
