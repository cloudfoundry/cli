package helpers

import (
	uuid "github.com/nu7hatch/gouuid"
)

func NewOrgName() string {
	return PrefixedRandomName("INTEGRATION-ORG")
}

func NewSpaceName() string {
	return PrefixedRandomName("INTEGRATION-SPACE")
}

func NewAppName() string {
	return PrefixedRandomName("INTEGRATION-APP")
}

func NewSecGroupName() string {
	return PrefixedRandomName("INTEGRATION-SEC-GROUP")
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
