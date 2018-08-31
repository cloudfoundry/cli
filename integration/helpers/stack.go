package helpers

import (
	"encoding/json"

	"fmt"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"strings"
)

type ccStacks struct {
	Resources []struct {
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	} `json:"resources"`
}

func FetchStacks() []string {
	session := CF("curl", "/v2/stacks")

	Eventually(session).Should(Exit(0))

	var rawStacks ccStacks
	err := json.Unmarshal(session.Out.Contents(), &rawStacks)
	Expect(err).ToNot(HaveOccurred())

	var stacks []string
	for _, stack := range rawStacks.Resources {
		stacks = append(stacks, stack.Entity.Name)
	}

	return stacks
}

func CreateStack(names ...string) string {
	name := NewStackName()
	if len(names) > 0 {
		name = names[0]
	}

	requestBody := fmt.Sprintf(
		`{"name":"%s","description":"CF CLI integration test stack, please delete"}`,
		name,
	)
	session := CF("curl", "-v", "-X", "POST", "/v2/stacks", "-d", requestBody)

	Eventually(session).Should(Say("201 Created"))
	Eventually(session).Should(Exit(0))

	return name
}

func DeleteStack(name string) {
	session := CF("stack", "--guid", name)
	Eventually(session).Should(Exit(0))
	guid := strings.TrimSpace(string(session.Out.Contents()))

	session = CF("curl", "-v", "-X", "DELETE", "/v2/stacks/"+guid)

	Eventually(session).Should(Say("204 No Content"))
	Eventually(session).Should(Exit(0))
}

func EnsureMinimumNumberOfStacks(num int) []string {
	var stacks []string
	for stacks = FetchStacks(); len(stacks) < 2; {
		stacks = append(stacks, CreateStack())
	}
	return stacks
}
