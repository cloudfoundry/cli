package api

import (
	"cf"
	"cf/net"
	"time"
	"net/http"
)

type FakeAppInstancesRepo struct{
	GetInstancesAppGuid    string
	GetInstancesResponses  [][]cf.AppInstanceFields
	GetInstancesErrorCodes []string
}

func (repo *FakeAppInstancesRepo) GetInstances(appGuid string) (instances[]cf.AppInstanceFields, apiResponse net.ApiResponse) {
	repo.GetInstancesAppGuid = appGuid
	time.Sleep(1*time.Millisecond) //needed for Windows only, otherwise it thinks error codes are not assigned

	if len(repo.GetInstancesResponses) > 0 {
		instances = repo.GetInstancesResponses[0]
		repo.GetInstancesResponses = repo.GetInstancesResponses[1:]
	}

	if len(repo.GetInstancesErrorCodes) > 0 {
		errorCode := repo.GetInstancesErrorCodes[0]
		repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]
		if errorCode != "" {
			apiResponse = net.NewApiResponse("Error staging app", errorCode, http.StatusBadRequest)
		}
	}

	return
}
