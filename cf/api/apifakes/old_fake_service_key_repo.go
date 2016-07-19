package apifakes

import (
	"code.cloudfoundry.org/cli/cf/models"
)

type OldFakeServiceKeyRepo struct {
	CreateServiceKeyMethod CreateServiceKeyType
	ListServiceKeysMethod  ListServiceKeysType
	GetServiceKeyMethod    GetServiceKeyType
	DeleteServiceKeyMethod DeleteServiceKeyType
}

type CreateServiceKeyType struct {
	InstanceGUID string
	KeyName      string
	Params       map[string]interface{}

	Error error
}

type ListServiceKeysType struct {
	InstanceGUID string

	ServiceKeys []models.ServiceKey
	Error       error
}

type GetServiceKeyType struct {
	InstanceGUID string
	KeyName      string

	ServiceKey models.ServiceKey
	Error      error
}

type DeleteServiceKeyType struct {
	GUID string

	Error error
}

func NewFakeServiceKeyRepo() *OldFakeServiceKeyRepo {
	return &OldFakeServiceKeyRepo{
		CreateServiceKeyMethod: CreateServiceKeyType{},
		ListServiceKeysMethod:  ListServiceKeysType{},
		GetServiceKeyMethod:    GetServiceKeyType{},
		DeleteServiceKeyMethod: DeleteServiceKeyType{},
	}
}

func (f *OldFakeServiceKeyRepo) CreateServiceKey(instanceGUID string, serviceKeyName string, params map[string]interface{}) error {
	f.CreateServiceKeyMethod.InstanceGUID = instanceGUID
	f.CreateServiceKeyMethod.KeyName = serviceKeyName
	f.CreateServiceKeyMethod.Params = params

	return f.CreateServiceKeyMethod.Error
}

func (f *OldFakeServiceKeyRepo) ListServiceKeys(instanceGUID string) ([]models.ServiceKey, error) {
	f.ListServiceKeysMethod.InstanceGUID = instanceGUID

	return f.ListServiceKeysMethod.ServiceKeys, f.ListServiceKeysMethod.Error
}

func (f *OldFakeServiceKeyRepo) GetServiceKey(instanceGUID string, serviceKeyName string) (models.ServiceKey, error) {
	f.GetServiceKeyMethod.InstanceGUID = instanceGUID

	return f.GetServiceKeyMethod.ServiceKey, f.GetServiceKeyMethod.Error
}

func (f *OldFakeServiceKeyRepo) DeleteServiceKey(serviceKeyGUID string) error {
	f.DeleteServiceKeyMethod.GUID = serviceKeyGUID

	return f.DeleteServiceKeyMethod.Error
}
