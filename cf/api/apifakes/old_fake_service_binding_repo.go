package apifakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type OldFakeServiceBindingRepo struct {
	CreateServiceInstanceGUID string
	CreateApplicationGUID     string
	CreateErrorCode           string
	CreateParams              map[string]interface{}

	DeleteServiceInstance models.ServiceInstance
	DeleteApplicationGUID string
	DeleteBindingNotFound bool
	CreateNonHTTPErrCode  string
}

func (repo *OldFakeServiceBindingRepo) Create(instanceGUID, appGUID string, paramsMap map[string]interface{}) (apiErr error) {
	repo.CreateServiceInstanceGUID = instanceGUID
	repo.CreateApplicationGUID = appGUID
	repo.CreateParams = paramsMap

	if repo.CreateNonHTTPErrCode != "" {
		apiErr = errors.New(repo.CreateNonHTTPErrCode)
		return
	}

	if repo.CreateErrorCode != "" {
		apiErr = errors.NewHTTPError(400, repo.CreateErrorCode, "Error binding service")
	}

	return
}

func (repo *OldFakeServiceBindingRepo) Delete(instance models.ServiceInstance, appGUID string) (found bool, apiErr error) {
	repo.DeleteServiceInstance = instance
	repo.DeleteApplicationGUID = appGUID
	found = !repo.DeleteBindingNotFound
	return
}
