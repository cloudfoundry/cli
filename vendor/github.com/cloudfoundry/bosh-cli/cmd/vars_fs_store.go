package cmd

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	cfgtypes "github.com/cloudfoundry/config-server/types"
	"gopkg.in/yaml.v2"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

type VarsFSStore struct {
	FS boshsys.FileSystem

	ValueGeneratorFactory cfgtypes.ValueGeneratorFactory

	path string
}

var _ boshtpl.Variables = VarsFSStore{}

func (s VarsFSStore) IsSet() bool { return len(s.path) > 0 }

func (s VarsFSStore) Get(varDef boshtpl.VariableDefinition) (interface{}, bool, error) {
	vars, err := s.load()
	if err != nil {
		return nil, false, err
	}

	val, found := vars[varDef.Name]
	if found {
		return val, true, nil
	}

	if len(varDef.Type) == 0 {
		return nil, false, nil
	}

	val, err = s.generateAndSet(varDef)
	if err != nil {
		return nil, false, bosherr.WrapErrorf(err, "Generating variable '%s'", varDef.Name)
	}

	return val, true, nil
}

func (s VarsFSStore) List() ([]boshtpl.VariableDefinition, error) {
	vars, err := s.load()
	if err != nil {
		return nil, err
	}

	return vars.List()
}

func (s VarsFSStore) generateAndSet(varDef boshtpl.VariableDefinition) (interface{}, error) {
	generator, err := s.ValueGeneratorFactory.GetGenerator(varDef.Type)
	if err != nil {
		return nil, err
	}

	val, err := generator.Generate(varDef.Options)
	if err != nil {
		return nil, err
	}

	err = s.set(varDef.Name, val)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (s VarsFSStore) set(key string, val interface{}) error {
	vars, err := s.load()
	if err != nil {
		return err
	}

	vars[key] = val

	return s.save(vars)
}

func (s VarsFSStore) load() (boshtpl.StaticVariables, error) {
	vars := boshtpl.StaticVariables{}

	if s.FS.FileExists(s.path) {
		bytes, err := s.FS.ReadFile(s.path)
		if err != nil {
			return vars, err
		}

		err = yaml.Unmarshal(bytes, &vars)
		if err != nil {
			return vars, bosherr.WrapErrorf(err, "Deserializing variables file store '%s'", s.path)
		}
	}

	return vars, nil
}

func (s VarsFSStore) save(vars boshtpl.StaticVariables) error {
	bytes, err := yaml.Marshal(vars)
	if err != nil {
		return bosherr.WrapErrorf(err, "Serializing variables")
	}

	err = s.FS.WriteFile(s.path, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing variables to file store '%s'", s.path)
	}

	return nil
}

func (s *VarsFSStore) UnmarshalFlag(data string) error {
	if len(data) == 0 {
		return bosherr.Errorf("Expected file path to be non-empty")
	}

	absPath, err := s.FS.ExpandPath(data)
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting absolute path '%s'", data)
	}

	(*s).path = absPath
	(*s).ValueGeneratorFactory = cfgtypes.NewValueGeneratorConcrete(nil)

	return nil
}
