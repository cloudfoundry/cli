package fakes

type FakeServiceKeyRepo struct {
	CreateServiceKeyError error

	CreateServiceKeyArgs CreateServiceKeyArgsType
}

type CreateServiceKeyArgsType struct {
	ServiceInstanceId string
	ServiceKeyName    string
}

func NewFakeServiceKeyRepo() *FakeServiceKeyRepo {
	return &FakeServiceKeyRepo{
		CreateServiceKeyArgs: CreateServiceKeyArgsType{},
	}
}

func (f *FakeServiceKeyRepo) CreateServiceKey(instanceId string, serviceKeyName string) (apiErr error) {
	f.CreateServiceKeyArgs.ServiceInstanceId = instanceId
	f.CreateServiceKeyArgs.ServiceKeyName = serviceKeyName

	return f.CreateServiceKeyError
}
