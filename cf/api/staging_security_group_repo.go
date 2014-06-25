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
	RemoveFromDefaultStagingSet(string) error
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
	return repo.gateway.UpdateResourceFromStruct(path, "")
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

//Delete /v2/config/staging_security_groups/a0bd0e1c-8609-45a9-88eb-f719b2e4367c
//204 No Content
func (repo *cloudControllerStagingSecurityGroupRepo) RemoveFromDefaultStagingSet(groupGuid string) error {
	path := fmt.Sprintf("%s/v2/config/staging_security_groups/%s", repo.configRepo.ApiEndpoint(), groupGuid)
	return repo.gateway.DeleteResource(path)
}
