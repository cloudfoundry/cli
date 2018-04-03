package stemcellfakes

import (
	"fmt"

	bicloud "github.com/cloudfoundry/bosh-cli/cloud"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	bitestutils "github.com/cloudfoundry/bosh-cli/testutils"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type NewManagerInput struct {
	Cloud bicloud.Cloud
}

type newManagerOutput struct {
	manager bistemcell.Manager
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

func (f *FakeManagerFactory) NewManager(cloud bicloud.Cloud) bistemcell.Manager {
	input := NewManagerInput{
		Cloud: cloud,
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

func (f *FakeManagerFactory) SetNewManagerBehavior(cloud bicloud.Cloud, manager bistemcell.Manager) {
	input := NewManagerInput{
		Cloud: cloud,
	}

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		panic(bosherr.WrapError(marshalErr, "Marshaling NewManager input"))
	}

	f.newManagerBehavior[inputString] = newManagerOutput{manager: manager}
}
