package fakes

import (
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeServiceKeyRepo struct {
	CreateServiceKeyMethod CreateServiceKeyType
	ListServiceKeysMethod  ListServiceKeysType
}

type CreateServiceKeyType struct {
	InstanceId string
	KeyName    string

	Error error
}

type ListServiceKeysType struct {
	InstanceId string

	ServiceKeys []models.ServiceKey
	Error       error
}

func NewFakeServiceKeyRepo() *FakeServiceKeyRepo {
	return &FakeServiceKeyRepo{
		CreateServiceKeyMethod: CreateServiceKeyType{},
		ListServiceKeysMethod:  ListServiceKeysType{},
	}
}

func (f *FakeServiceKeyRepo) CreateServiceKey(instanceId string, serviceKeyName string) error {
	f.CreateServiceKeyMethod.InstanceId = instanceId
	f.CreateServiceKeyMethod.KeyName = serviceKeyName

	return f.CreateServiceKeyMethod.Error
}

func (f *FakeServiceKeyRepo) ListServiceKeys(instanceId string) ([]models.ServiceKey, error) {
	f.ListServiceKeysMethod.InstanceId = instanceId

	return f.ListServiceKeysMethod.ServiceKeys, f.ListServiceKeysMethod.Error
}
