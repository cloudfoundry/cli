package configuration

import (
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
	Load() (DataInterface, error)
	Save(DataInterface) error
}

type DataInterface interface {
	NewData() DataInterface
	JsonMarshalV3() ([]byte, error)
	JsonUnmarshalV3([]byte) error
}

type DiskPersistor struct {
	data     DataInterface
	filePath string
}

func NewDiskPersistor(path string, data DataInterface) (dp DiskPersistor) {
	return DiskPersistor{
		filePath: path,
		data:     data,
	}
}

func (dp DiskPersistor) Delete() {
	os.Remove(dp.filePath)
}

func (dp DiskPersistor) Load() (data DataInterface, err error) {
	data, err = dp.read()
	if err != nil {
		err = dp.write(data)
	}
	return
}

func (dp DiskPersistor) Save(data DataInterface) (err error) {
	return dp.write(data)
}

func (dp DiskPersistor) read() (data DataInterface, err error) {
	data = dp.data.NewData()

	err = os.MkdirAll(filepath.Dir(dp.filePath), dirPermissions)
	if err != nil {
		return
	}

	jsonBytes, err := ioutil.ReadFile(dp.filePath)
	if err != nil {
		return
	}

	err = data.JsonUnmarshalV3(jsonBytes)
	return
}

func (dp DiskPersistor) write(data DataInterface) (err error) {
	bytes, err := data.JsonMarshalV3()
	if err != nil {
		return
	}

	err = ioutil.WriteFile(dp.filePath, bytes, filePermissions)
	return
}
