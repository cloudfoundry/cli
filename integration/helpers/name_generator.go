package helpers

import (
	uuid "github.com/nu7hatch/gouuid"
)

func NewOrgName() string {
	return PrefixedRandomName("ORG")
}

func PrefixedRandomName(namePrefix string) string {
	return namePrefix + "-" + RandomName()
}

func RandomName() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return guid.String()
}
