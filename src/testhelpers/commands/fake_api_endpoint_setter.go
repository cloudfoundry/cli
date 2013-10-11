package commands

type FakeApiEndpointSetter struct {
	SetEndpoint string
}

func (setter *FakeApiEndpointSetter) SetApiEndpoint(endpoint string) {
	setter.SetEndpoint = endpoint
	return
}
