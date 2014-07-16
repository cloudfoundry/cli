package configuration

import (
	"errors"
	"fmt"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	filePermissions = 0600
	dirPermissions  = 0700
)

type Persistor interface {
	Delete()
	Load() (*Data, error)
	Save(*Data) error
}

type DiskPersistor struct {
	filePath string
}

func NewDiskPersistor(path string) (dp DiskPersistor) {
	return DiskPersistor{filePath: path}
}

func (dp DiskPersistor) Delete() {
	os.Remove(dp.filePath)
}

func (dp DiskPersistor) Load() (data *Data, err error) {
	data, err = dp.read()
	if err != nil {
		err = dp.write(data)
	}
	return
}

func (dp DiskPersistor) Save(data *Data) (err error) {
	return dp.write(data)
}

func (dp DiskPersistor) read() (data *Data, err error) {
	data = NewData()

	err = os.MkdirAll(filepath.Dir(dp.filePath), dirPermissions)
	if err != nil {
		return
	}

	jsonBytes, err := ioutil.ReadFile(dp.filePath)
	if err != nil {
		return
	}

	err = JsonUnmarshalV3(jsonBytes, data)
	return
}

func (dp DiskPersistor) write(data *Data) (err error) {
	bytes, err := JsonMarshalV3(data)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(dp.filePath, bytes, filePermissions)
	if err != nil {
		err = errors.New(fmt.Sprintf(T("Error writing to manifest file:{{.FilePath}}\n{{.Err}}",
			map[string]interface{}{"FilePath": dp.filePath, "Err": err})))
		return
	}
	return
}
