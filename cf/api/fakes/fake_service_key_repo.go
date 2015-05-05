package fakes

import (
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeServiceKeyRepo struct {
	CreateServiceKeyMethod CreateServiceKeyType
	ListServiceKeysMethod  ListServiceKeysType
	GetServiceKeyMethod    GetServiceKeyType
	DeleteServiceKeyMethod DeleteServiceKeyType
}

type CreateServiceKeyType struct {
	InstanceGuid string
	KeyName      string

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

func NewFakeServiceKeyRepo() *FakeServiceKeyRepo {
	return &FakeServiceKeyRepo{
		CreateServiceKeyMethod: CreateServiceKeyType{},
		ListServiceKeysMethod:  ListServiceKeysType{},
		GetServiceKeyMethod:    GetServiceKeyType{},
		DeleteServiceKeyMethod: DeleteServiceKeyType{},
	}
}

func (f *FakeServiceKeyRepo) CreateServiceKey(instanceGuid string, serviceKeyName string) error {
	f.CreateServiceKeyMethod.InstanceGuid = instanceGuid
	f.CreateServiceKeyMethod.KeyName = serviceKeyName

	return f.CreateServiceKeyMethod.Error
}

func (f *FakeServiceKeyRepo) ListServiceKeys(instanceGuid string) ([]models.ServiceKey, error) {
	f.ListServiceKeysMethod.InstanceGuid = instanceGuid

	return f.ListServiceKeysMethod.ServiceKeys, f.ListServiceKeysMethod.Error
}

func (f *FakeServiceKeyRepo) GetServiceKey(instanceGuid string, serviceKeyName string) (models.ServiceKey, error) {
	f.GetServiceKeyMethod.InstanceGuid = instanceGuid

	return f.GetServiceKeyMethod.ServiceKey, f.GetServiceKeyMethod.Error
}

func (f *FakeServiceKeyRepo) DeleteServiceKey(serviceKeyGuid string) error {
	f.DeleteServiceKeyMethod.Guid = serviceKeyGuid

	return f.DeleteServiceKeyMethod.Error
}
