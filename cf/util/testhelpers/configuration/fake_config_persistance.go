package configuration

import (
	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
)

type FakePersistor struct {
	LoadReturns struct {
		Data *coreconfig.Data
		Err  error
	}

	SaveArgs struct {
		Data *coreconfig.Data
	}
	SaveReturns struct {
		Err error
	}
}

func NewFakePersistor() *FakePersistor {
	return &FakePersistor{}
}

func (fp *FakePersistor) Load(data configuration.DataInterface) error {
	if fp.LoadReturns.Data == nil {
		fp.LoadReturns.Data = coreconfig.NewData()
	}
	return fp.LoadReturns.Err
}

func (fp *FakePersistor) Delete() {}

func (fp *FakePersistor) Exists() bool {
	return true
}

func (fp *FakePersistor) Save(data configuration.DataInterface) (err error) {
	fp.SaveArgs.Data = data.(*coreconfig.Data)
	err = fp.SaveReturns.Err
	return
}
