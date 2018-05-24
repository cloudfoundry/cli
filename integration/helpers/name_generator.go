package helpers

import (
	uuid "github.com/nu7hatch/gouuid"
)

func NewAppName() string {
	return PrefixedRandomName("INTEGRATION-APP")
}

func NewIsolationSegmentName(infix ...string) string {
	return PrefixedRandomName("INTEGRATION-ISOLATION-SEGMENT")
}

func NewOrgName() string {
	return PrefixedRandomName("INTEGRATION-ORG")
}

func NewServiceBrokerName() string {
	return PrefixedRandomName("INTEGRATION-SERVICE-BROKER")
}

func NewPassword() string {
	return PrefixedRandomName("INTEGRATION-PASSWORD")
}

func NewSecurityGroupName(infix ...string) string {
	if len(infix) > 0 {
		return PrefixedRandomName("INTEGRATION-SEC-GROUP-" + infix[0])
	}

	return PrefixedRandomName("INTEGRATION-SEC-GROUP")
}

func NewSpaceName() string {
	return PrefixedRandomName("INTEGRATION-SPACE")
}

func NewUsername() string {
	return PrefixedRandomName("INTEGRATION-USER")
}

func NewBuildpack() string {
	return PrefixedRandomName("INTEGRATION-BUILDPACK")
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
