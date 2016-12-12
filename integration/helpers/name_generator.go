package helpers

import (
	uuid "github.com/nu7hatch/gouuid"
)

func NewOrgName() string {
	return PrefixedRandomName("INTEGRATION-ORG")
}

func RandomUsername() string {
	return PrefixedRandomName("integration-user")
}

func RandomPassword() string {
	return PrefixedRandomName("password")
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
