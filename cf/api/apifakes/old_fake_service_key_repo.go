package apifakes

import (
	"github.com/cloudfoundry/cli/cf/models"
)

type OldFakeServiceKeyRepo struct {
	CreateServiceKeyMethod CreateServiceKeyType
	ListServiceKeysMethod  ListServiceKeysType
	GetServiceKeyMethod    GetServiceKeyType
	DeleteServiceKeyMethod DeleteServiceKeyType
}

type CreateServiceKeyType struct {
	InstanceGuid string
	KeyName      string
	Params       map[string]interface{}

	Error error
}

type ListServiceKeysType struct {
	InstanceGuid string

	ServiceKeys []models.ServiceKey
	Error       error
}

type GetServiceKeyType struct {
	InstanceGuid string
	KeyName      string

	ServiceKey models.ServiceKey
	Error      error
}

type DeleteServiceKeyType struct {
	Guid string

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

func (f *OldFakeServiceKeyRepo) CreateServiceKey(instanceGuid string, serviceKeyName string, params map[string]interface{}) error {
	f.CreateServiceKeyMethod.InstanceGuid = instanceGuid
	f.CreateServiceKeyMethod.KeyName = serviceKeyName
	f.CreateServiceKeyMethod.Params = params

	return f.CreateServiceKeyMethod.Error
}

func (f *OldFakeServiceKeyRepo) ListServiceKeys(instanceGuid string) ([]models.ServiceKey, error) {
	f.ListServiceKeysMethod.InstanceGuid = instanceGuid

	return f.ListServiceKeysMethod.ServiceKeys, f.ListServiceKeysMethod.Error
}

func (f *OldFakeServiceKeyRepo) GetServiceKey(instanceGuid string, serviceKeyName string) (models.ServiceKey, error) {
	f.GetServiceKeyMethod.InstanceGuid = instanceGuid

	return f.GetServiceKeyMethod.ServiceKey, f.GetServiceKeyMethod.Error
}

func (f *OldFakeServiceKeyRepo) DeleteServiceKey(serviceKeyGuid string) error {
	f.DeleteServiceKeyMethod.Guid = serviceKeyGuid

	return f.DeleteServiceKeyMethod.Error
}
