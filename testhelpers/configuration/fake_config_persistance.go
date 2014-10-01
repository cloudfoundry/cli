package configuration

import (
	"github.com/cloudfoundry/cli/cf/configuration"
)

type FakePersistor struct {
	LoadReturns struct {
		Data configuration.DataInterface
		Err  error
	}

	SaveArgs struct {
		Data configuration.DataInterface
	}
	SaveReturns struct {
		Err error
	}
}

func NewFakePersistor() *FakePersistor {
	return &FakePersistor{}
}

func (fp *FakePersistor) Load() (data configuration.DataInterface, err error) {
	if fp.LoadReturns.Data == nil {
		fp.LoadReturns.Data = configuration.NewData()
	}
	data = fp.LoadReturns.Data
	err = fp.LoadReturns.Err
	return
}

func (fp *FakePersistor) Delete() {

}

func (fp *FakePersistor) Save(data configuration.DataInterface) (err error) {
	fp.SaveArgs.Data = data
	err = fp.SaveReturns.Err
	return
}
