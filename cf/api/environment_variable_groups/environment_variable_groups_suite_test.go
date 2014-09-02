package environment_variable_groups_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEnvironmentVariableGroups(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EnvironmentVariableGroups Suite")
}
