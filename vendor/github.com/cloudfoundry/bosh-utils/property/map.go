package property

type Map map[string]Property

func (m *Map) UnmarshalYAML(unmarshal func(interface{}) error) error {
	rawMap := map[interface{}]interface{}{}
	err := unmarshal(&rawMap)
	if err != nil {
		return err
	}

	*m, err = BuildMap(rawMap)
	if err != nil {
		return err
	}

	return nil
}
