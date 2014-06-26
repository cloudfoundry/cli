package defaults

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type Methods interface {
	Add(string, string) error
	List(string) ([]models.SecurityGroupFields, error)
}

type SecurityGroupsRepoBase struct {
	ConfigRepo configuration.Reader
	Gateway    net.Gateway
}

func (repo *SecurityGroupsRepoBase) Add(groupGuid string, path string) error {
	updatedPath := fmt.Sprintf("%s%s%s", repo.ConfigRepo.ApiEndpoint(), path, groupGuid)
	return repo.Gateway.UpdateResourceFromStruct(updatedPath, "")
}

func (repo *SecurityGroupsRepoBase) List(path string) ([]models.SecurityGroupFields, error) {
	groups := []models.SecurityGroupFields{}

	err := repo.Gateway.ListPaginatedResources(
		repo.ConfigRepo.ApiEndpoint(),
		path,
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
