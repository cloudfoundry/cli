package security_groups

import (
	"fmt"
	"net/url"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type SecurityGroupRepo interface {
	Create(name string, rules []map[string]interface{}) error
	Update(guid string, rules []map[string]interface{}) error
	Read(string) (models.SecurityGroup, error)
	Delete(string) error
	FindAll() ([]models.SecurityGroup, error)
}

type cloudControllerSecurityGroupRepo struct {
	gateway net.Gateway
	config  core_config.Reader
}

func NewSecurityGroupRepo(config core_config.Reader, gateway net.Gateway) SecurityGroupRepo {
	return cloudControllerSecurityGroupRepo{
		config:  config,
		gateway: gateway,
	}
}

func (repo cloudControllerSecurityGroupRepo) Create(name string, rules []map[string]interface{}) error {
	path := "/v2/security_groups"
	params := models.SecurityGroupParams{
		Name:  name,
		Rules: rules,
	}
	return repo.gateway.CreateResourceFromStruct(repo.config.ApiEndpoint(), path, params)
}

func (repo cloudControllerSecurityGroupRepo) Read(name string) (models.SecurityGroup, error) {
	path := fmt.Sprintf("/v2/security_groups?q=%s", url.QueryEscape("name:"+name))
	group := models.SecurityGroup{}
	foundGroup := false

	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.SecurityGroupResource{},
		func(resource interface{}) bool {
			if asgr, ok := resource.(resources.SecurityGroupResource); ok {
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
		return group, errors.NewModelNotFoundError("security group", name)
	}

	err = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		group.SpaceUrl+"?inline-relations-depth=1",
		resources.SpaceResource{},
		func(resource interface{}) bool {
			if asgr, ok := resource.(resources.SpaceResource); ok {
				group.Spaces = append(group.Spaces, asgr.ToModel())
				return true
			}
			return false
		},
	)

	return group, err
}

func (repo cloudControllerSecurityGroupRepo) Update(guid string, rules []map[string]interface{}) error {
	url := fmt.Sprintf("/v2/security_groups/%s", guid)
	return repo.gateway.UpdateResourceFromStruct(repo.config.ApiEndpoint(), url, models.SecurityGroupParams{Rules: rules})
}

func (repo cloudControllerSecurityGroupRepo) FindAll() ([]models.SecurityGroup, error) {
	path := "/v2/security_groups"
	securityGroups := []models.SecurityGroup{}

	err := repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		path,
		resources.SecurityGroupResource{},
		func(resource interface{}) bool {
			if securityGroupResource, ok := resource.(resources.SecurityGroupResource); ok {
				securityGroups = append(securityGroups, securityGroupResource.ToModel())
			}

			return true
		},
	)

	if err != nil {
		return nil, err
	}

	for i := range securityGroups {
		err = repo.gateway.ListPaginatedResources(
			repo.config.ApiEndpoint(),
			securityGroups[i].SpaceUrl+"?inline-relations-depth=1",
			resources.SpaceResource{},
			func(resource interface{}) bool {
				if asgr, ok := resource.(resources.SpaceResource); ok {
					securityGroups[i].Spaces = append(securityGroups[i].Spaces, asgr.ToModel())
					return true
				}
				return false
			},
		)
	}

	return securityGroups, err
}

func (repo cloudControllerSecurityGroupRepo) Delete(securityGroupGuid string) error {
	path := fmt.Sprintf("/v2/security_groups/%s", securityGroupGuid)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}
