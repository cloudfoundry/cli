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
	Exists() bool
	Load(DataInterface) error
	Save(DataInterface) error
}

type DataInterface interface {
	JsonMarshalV3() ([]byte, error)
	JsonUnmarshalV3([]byte) error
}

type DiskPersistor struct {
	filePath string
}

func NewDiskPersistor(path string) (dp DiskPersistor) {
	return DiskPersistor{
		filePath: path,
	}
}

func (dp DiskPersistor) Exists() bool {
	_, err := os.Stat(dp.filePath)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}

func (dp DiskPersistor) Delete() {
	os.Remove(dp.filePath)
}

func (dp DiskPersistor) Load(data DataInterface) error {
	err := dp.read(data)
	if os.IsPermission(err) {
		return err
	}

	if err != nil {
		err = dp.write(data)
	}
	return err
}

func (dp DiskPersistor) Save(data DataInterface) (err error) {
	return dp.write(data)
}

func (dp DiskPersistor) read(data DataInterface) error {
	err := os.MkdirAll(filepath.Dir(dp.filePath), dirPermissions)
	if err != nil {
		return err
	}

	jsonBytes, err := ioutil.ReadFile(dp.filePath)
	if err != nil {
		return err
	}

	err = data.JsonUnmarshalV3(jsonBytes)
	return err
}

func (dp DiskPersistor) write(data DataInterface) error {
	bytes, err := data.JsonMarshalV3()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dp.filePath, bytes, filePermissions)
	return err
}
