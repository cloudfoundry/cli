package api

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type StagingSecurityGroupsRepo interface {
	AddToDefaultStagingSet(string) error
	List() ([]models.SecurityGroupFields, error)
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

func (repo *cloudControllerStagingSecurityGroupRepo) List() ([]models.SecurityGroupFields, error) {
	groups := []models.SecurityGroupFields{}

	err := repo.gateway.ListPaginatedResources(
		repo.configRepo.ApiEndpoint(),
		"/v2/config/staging_security_groups",
		resources.SecurityGroupResource{},
		func(resource interface{}) bool {
			if securityGroupResource, ok := resource.(resources.SecurityGroupResource); ok {
				groups = append(groups, securityGroupResource.ToFields())
			}

			return true
		},
	)

	return groups, err
}
