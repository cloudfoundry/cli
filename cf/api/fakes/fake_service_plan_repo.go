package fakes

import (
	"sort"
	"strings"

	"github.com/cloudfoundry/cli/cf/models"
)

type FakeServicePlanRepo struct {
	SearchReturns map[string][]models.ServicePlanFields
	SearchErr     error
}

func (fake FakeServicePlanRepo) Search(queryParams map[string]string) ([]models.ServicePlanFields, error) {
	if fake.SearchErr != nil {
		return nil, fake.SearchErr
	}

	if queryParams == nil {
		//return everything
		var returnPlans []models.ServicePlanFields
		for _, value := range fake.SearchReturns {
			returnPlans = append(returnPlans, value...)
		}
		return returnPlans, nil
	}

	searchKey := combineKeys(queryParams)
	if fake.SearchReturns[searchKey] != nil {
		return fake.SearchReturns[searchKey], nil
	}

	return []models.ServicePlanFields{}, nil
}

func combineKeys(mapToCombine map[string]string) string {
	keys := []string{}
	for key, _ := range mapToCombine {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	values := []string{}
	for _, key := range keys {
		values = append(values, mapToCombine[key])
	}

	return strings.Join(values, ":")
}
