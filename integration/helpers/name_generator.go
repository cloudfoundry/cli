package helpers

import (
	uuid "github.com/nu7hatch/gouuid"
)

func RandomName() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return guid.String()
}

func PrefixedRandomName(namePrefix string) string {
	return namePrefix + "-" + RandomName()
}
