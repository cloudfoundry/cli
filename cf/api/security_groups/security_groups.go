package security_groups

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
	Create(name string, rules []map[string]interface{}) error
	Update(guid string, rules []map[string]interface{}) error
	Read(string) (models.SecurityGroup, error)
	Delete(string) error
	FindAll() ([]models.SecurityGroup, error)
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

func (repo cloudControllerSecurityGroupRepo) Create(name string, rules []map[string]interface{}) error {
	path := fmt.Sprintf("%s/v2/security_groups", repo.config.ApiEndpoint())
	params := models.SecurityGroupParams{
		Name:  name,
		Rules: rules,
	}
	return repo.gateway.CreateResourceFromStruct(path, params)
}

func (repo cloudControllerSecurityGroupRepo) Read(name string) (models.SecurityGroup, error) {
	path := fmt.Sprintf("/v2/security_groups?q=%s&inline-relations-depth=2", url.QueryEscape("name:"+name))
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
		err = errors.NewModelNotFoundError("security group", name)
	}

	return group, err
}

func (repo cloudControllerSecurityGroupRepo) Update(guid string, rules []map[string]interface{}) error {
	url := fmt.Sprintf("%s/v2/security_groups/%s", repo.config.ApiEndpoint(), guid)
	return repo.gateway.UpdateResourceFromStruct(url, models.SecurityGroupParams{Rules: rules})
}

func (repo cloudControllerSecurityGroupRepo) FindAll() ([]models.SecurityGroup, error) {
	path := "/v2/security_groups?inline-relations-depth=2"
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

	return securityGroups, err
}

func (repo cloudControllerSecurityGroupRepo) Delete(securityGroupGuid string) error {
	path := fmt.Sprintf("%s/v2/security_groups/%s", repo.config.ApiEndpoint(), securityGroupGuid)
	return repo.gateway.DeleteResource(path)
}
