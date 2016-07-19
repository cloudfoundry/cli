package apifakes

import (
	"sort"
	"strings"
	"sync"

	"code.cloudfoundry.org/cli/cf/models"
)

type OldFakeServicePlanRepo struct {
	SearchReturns map[string][]models.ServicePlanFields
	SearchErr     error

	UpdateStub        func(models.ServicePlanFields, string, bool) error
	updateMutex       sync.RWMutex
	updateArgsForCall []struct {
		arg1 models.ServicePlanFields
		arg2 string
		arg3 bool
	}
	updateReturns struct {
		result1 error
	}

	ListPlansFromManyServicesReturns []models.ServicePlanFields
	ListPlansFromManyServicesError   error
}

func (fake *OldFakeServicePlanRepo) ListPlansFromManyServices(serviceGUIDs []string) (plans []models.ServicePlanFields, err error) {
	if fake.ListPlansFromManyServicesError != nil {
		return nil, fake.ListPlansFromManyServicesError
	}

	if fake.ListPlansFromManyServicesReturns != nil {
		return fake.ListPlansFromManyServicesReturns, nil
	}
	return []models.ServicePlanFields{}, nil
}

func (fake *OldFakeServicePlanRepo) Search(queryParams map[string]string) ([]models.ServicePlanFields, error) {
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
	for key := range mapToCombine {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	values := []string{}
	for _, key := range keys {
		values = append(values, mapToCombine[key])
	}

	return strings.Join(values, ":")
}

func (fake *OldFakeServicePlanRepo) Update(arg1 models.ServicePlanFields, arg2 string, arg3 bool) error {
	fake.updateMutex.Lock()
	defer fake.updateMutex.Unlock()
	fake.updateArgsForCall = append(fake.updateArgsForCall, struct {
		arg1 models.ServicePlanFields
		arg2 string
		arg3 bool
	}{arg1, arg2, arg3})
	if fake.UpdateStub != nil {
		return fake.UpdateStub(arg1, arg2, arg3)
	} else {
		return fake.updateReturns.result1
	}
}

func (fake *OldFakeServicePlanRepo) UpdateCallCount() int {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	return len(fake.updateArgsForCall)
}

func (fake *OldFakeServicePlanRepo) UpdateArgsForCall(i int) (models.ServicePlanFields, string, bool) {
	fake.updateMutex.RLock()
	defer fake.updateMutex.RUnlock()
	return fake.updateArgsForCall[i].arg1, fake.updateArgsForCall[i].arg2, fake.updateArgsForCall[i].arg3
}

func (fake *OldFakeServicePlanRepo) UpdateReturns(result1 error) {
	fake.UpdateStub = nil
	fake.updateReturns = struct {
		result1 error
	}{result1}
}
