package fakes

import (
	"fmt"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bivm "github.com/cloudfoundry/bosh-cli/deployment/vm"
	bitestutils "github.com/cloudfoundry/bosh-cli/testutils"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type NewManagerInput struct {
	Cloud   bicloud.Cloud
	MbusURL string
}

type newManagerOutput struct {
	manager bivm.Manager
}

type FakeManagerFactory struct {
	NewManagerInputs   []NewManagerInput
	newManagerBehavior map[string]newManagerOutput
}

func NewFakeManagerFactory() *FakeManagerFactory {
	return &FakeManagerFactory{
		NewManagerInputs:   []NewManagerInput{},
		newManagerBehavior: map[string]newManagerOutput{},
	}
}

func (f *FakeManagerFactory) NewManager(cloud bicloud.Cloud, mbusURL string) bivm.Manager {
	input := NewManagerInput{
		Cloud:   cloud,
		MbusURL: mbusURL,
	}
	f.NewManagerInputs = append(f.NewManagerInputs, input)

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	output, found := f.newManagerBehavior[inputString]
	if !found {
		panic(fmt.Errorf("Unsupported NewManager Input: %#v\nExpected Behavior: %#v", input, f.newManagerBehavior))
	}

	return output.manager
}

func (f *FakeManagerFactory) SetNewManagerBehavior(cloud bicloud.Cloud, mbusURL string, manager bivm.Manager) {
	input := NewManagerInput{
		Cloud:   cloud,
		MbusURL: mbusURL,
	}

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	f.newManagerBehavior[inputString] = newManagerOutput{manager: manager}
}
