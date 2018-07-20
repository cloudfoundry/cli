package helpers

import (
	"sort"

	uuid "github.com/nu7hatch/gouuid"
)

// TODO: Is this working???
func GenerateHigherName(randomNameGenerator func() string, names ...string) string {
	name := randomNameGenerator()

	var safe bool
	for !safe {
		safe = true
		for _, nameToBeLowerThan := range names {
			// regenerate name if name is NOT higher
			if !sort.StringsAreSorted([]string{nameToBeLowerThan, name}) {
				name = randomNameGenerator()
				safe = false
				break
			}
		}
	}
	return name
}

// TODO: Is this working???
func GenerateLowerName(randomNameGenerator func() string, names ...string) string {
	name := randomNameGenerator()

	var safe bool
	for !safe {
		safe = true
		for _, nameToBeHigherThan := range names {
			// regenerate name if name is NOT lower
			if !sort.StringsAreSorted([]string{name, nameToBeHigherThan}) {
				name = randomNameGenerator()
				safe = false
				break
			}
		}
	}
	return name
}

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

func NewPlanName() string {
	return PrefixedRandomName("INTEGRATION-PLAN")
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
