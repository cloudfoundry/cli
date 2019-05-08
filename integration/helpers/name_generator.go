package helpers

import (
	"fmt"
	"sort"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
)

// TODO: Is this working???

// GenerateHigherName will use the passed randomNameGenerator to generate a name with a higher
// sort value than all the passed names
func GenerateHigherName(randomNameGenerator func() string, names ...string) string {
	if len(names) == 0 {
		return randomNameGenerator()
	}
	sort.Strings(names)

	maxName := names[len(names)-1]

	for {
		name := randomNameGenerator()
		// regenerate name if name is NOT higher
		if strings.Compare(name, maxName) > 0 {
			return name
		}
	}
}

// TODO: Is this working???

// GenerateLowerName will use the passed randomNameGenerator to generate a name with a lower
// sort value than all the passed names
func GenerateLowerName(randomNameGenerator func() string, names ...string) string {
	if len(names) == 0 {
		return randomNameGenerator()
	}

	sort.Strings(names)

	minName := names[0]

	for {
		name := randomNameGenerator()
		// regenerate name if name is NOT higher
		if strings.Compare(name, minName) < 0 {
			return name
		}
	}
}

// NewAppName provides a random name prefixed with INTEGRATION-APP
func NewAppName() string {
	return PrefixedRandomName("INTEGRATION-APP")
}

// NewIsolationSegmentName provides a random name prefixed with INTEGRATION-ISOLATION-SEGMENT
func NewIsolationSegmentName(infix ...string) string {
	return PrefixedRandomName("INTEGRATION-ISOLATION-SEGMENT")
}

// NewOrgName provides a random name prefixed with INTEGRATION-ORG
func NewOrgName() string {
	return PrefixedRandomName("INTEGRATION-ORG")
}

// NewServiceBrokerName provides a random name prefixed with INTEGRATION-SERVICE-BROKER
func NewServiceBrokerName() string {
	return PrefixedRandomName("INTEGRATION-SERVICE-BROKER")
}

// NewPlanName provides a random name prefixed with INTEGRATION-PLAN
func NewPlanName() string {
	return PrefixedRandomName("INTEGRATION-PLAN")
}

// NewPassword provides a random string prefixed with INTEGRATION-PASSWORD
func NewPassword() string {
	return PrefixedRandomName("INTEGRATION-PASSWORD")
}

// NewSecurityGroupName provides a random name prefixed with INTEGRATION-SEC-GROUP. If an infix is provided, it
// is placed between INTEGRATION-SEC-GROUP and the random string.
func NewSecurityGroupName(infix ...string) string {
	if len(infix) > 0 {
		return PrefixedRandomName("INTEGRATION-SEC-GROUP-" + infix[0])
	}

	return PrefixedRandomName("INTEGRATION-SEC-GROUP")
}

// NewSpaceName provides a random name prefixed with INTEGRATION-SPACE
func NewSpaceName() string {
	return PrefixedRandomName("INTEGRATION-SPACE")
}

// NewUsername provides a random name prefixed with INTEGRATION-USER
func NewUsername() string {
	return PrefixedRandomName("INTEGRATION-USER")
}

// NewBuildpackName provides a random name prefixed with INTEGRATION-BUILDPACK
func NewBuildpackName() string {
	return PrefixedRandomName("INTEGRATION-BUILDPACK")
}

// NewStackName provides a random name prefixed with INTEGRATION-STACK
func NewStackName() string {
	return PrefixedRandomName("INTEGRATION-STACK")
}

// NewDomainName provides a random domain name prefixed with integration. If prefix is provided the domain name
// will have structure "integration-prefix-randomstring.com" else it will have structure "integration-randomstring.com"
func NewDomainName(prefix ...string) string {
	if len(prefix) > 0 {
		return fmt.Sprintf("integration-%s.com", PrefixedRandomName(prefix[0]))
	}
	return fmt.Sprintf("integration%s.com", PrefixedRandomName(""))
}

// PrefixedRandomName provides a random name with structure "namePrefix-randomstring"
func PrefixedRandomName(namePrefix string) string {
	return namePrefix + "-" + RandomName()
}

// RandomName provides a random string
func RandomName() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return guid.String()
}
