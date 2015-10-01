package fakes

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/models"
)

type FakeRoutingApiRepository struct {
	RouterGroups models.RouterGroups
	ListError    bool
}

func (fake *FakeRoutingApiRepository) ListRouterGroups(cb func(models.RouterGroup) bool) (apiErr error) {
	if fake.ListError {
		apiErr = errors.New("Error in FakeRoutingApiRepository")
		return
	}

	for _, routerGroup := range fake.RouterGroups {
		if !cb(routerGroup) {
			break
		}
	}
	return
}
