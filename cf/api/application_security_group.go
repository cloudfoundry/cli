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

type SecurityGroupRepo interface {
	Create(name string, rules []map[string]string, spaceGuids []string) error
	Delete(string) error
	Read(string) (models.ApplicationSecurityGroup, error)
	FindAll() ([]models.ApplicationSecurityGroup, error)
}

type cloudControllerSecurityGroupRepo struct {
	gateway net.Gateway
	config  configuration.Reader
}

func NewSecurityGroupRepo(config configuration.Reader, gateway net.Gateway) SecurityGroupRepo {
	return cloudControllerSecurityGroupRepo{
		config:  config,
		gateway: gateway,
	}
}

func (repo cloudControllerSecurityGroupRepo) Create(name string, rules []map[string]string, spaceGuids []string) error {
	path := fmt.Sprintf("%s/v2/app_security_groups", repo.config.ApiEndpoint())
	params := models.ApplicationSecurityGroupParams{
		Name:       name,
		Rules:      rules,
		SpaceGuids: spaceGuids,
	}
	return repo.gateway.CreateResourceFromStruct(path, params)
}

func (repo cloudControllerSecurityGroupRepo) Read(name string) (models.ApplicationSecurityGroup, error) {
	path := fmt.Sprintf("/v2/app_security_groups?q=%s&inline-relations-depth=2", url.QueryEscape("name:"+name))
	group := models.ApplicationSecurityGroup{}
	foundGroup := false

	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.ApplicationSecurityGroupResource{},
		func(resource interface{}) bool {
			if asgr, ok := resource.(resources.ApplicationSecurityGroupResource); ok {
				group = asgr.ToModel()
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

func (repo cloudControllerSecurityGroupRepo) FindAll() ([]models.ApplicationSecurityGroup, error) {
	path := "/v2/app_security_groups?inline-relations-depth=2"
	securityGroups := []models.ApplicationSecurityGroup{}

	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.ApplicationSecurityGroupResource{},
		func(resource interface{}) bool {
			if securityGroupResource, ok := resource.(resources.ApplicationSecurityGroupResource); ok {
				securityGroups = append(securityGroups, securityGroupResource.ToModel())
			}

			return true
		},
	)

	if err != nil {
		return nil, err
	}

	return securityGroups, err
}

func (repo cloudControllerSecurityGroupRepo) Delete(securityGroupGuid string) error {
	path := fmt.Sprintf("%s/v2/app_security_groups/%s", repo.config.ApiEndpoint(), securityGroupGuid)
	return repo.gateway.DeleteResource(path)
}
