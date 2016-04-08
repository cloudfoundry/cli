package apifakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type OldFakeServiceBindingRepo struct {
	CreateServiceInstanceGuid string
	CreateApplicationGuid     string
	CreateErrorCode           string
	CreateParams              map[string]interface{}

	DeleteServiceInstance models.ServiceInstance
	DeleteApplicationGuid string
	DeleteBindingNotFound bool
	CreateNonHttpErrCode  string
}

func (repo *OldFakeServiceBindingRepo) Create(instanceGuid, appGuid string, paramsMap map[string]interface{}) (apiErr error) {
	repo.CreateServiceInstanceGuid = instanceGuid
	repo.CreateApplicationGuid = appGuid
	repo.CreateParams = paramsMap

	if repo.CreateNonHttpErrCode != "" {
		apiErr = errors.New(repo.CreateNonHttpErrCode)
		return
	}

	if repo.CreateErrorCode != "" {
		apiErr = errors.NewHttpError(400, repo.CreateErrorCode, "Error binding service")
	}

	return
}

func (repo *OldFakeServiceBindingRepo) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr error) {
	repo.DeleteServiceInstance = instance
	repo.DeleteApplicationGuid = appGuid
	found = !repo.DeleteBindingNotFound
	return
}
