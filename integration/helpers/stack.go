package helpers

import (
	"encoding/json"

	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

type ccStacks struct {
	Resources []struct {
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	} `json:"resources"`
}

// FetchStacks returns all the stack names present in the foundation.
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

// PreferredStack returns the cflinuxfs3 stack name if it present, otherwise cflinuxfs2 is returned.
func PreferredStack() string {
	stacks := FetchStacks()

	for _, name := range stacks {
		if name == "cflinuxfs3" {
			return name
		}
	}

	return "cflinuxfs2"
}

// CreateStack creates a new stack with the user provided name. If a name is not provided, a random name is used
func CreateStack(names ...string) string {
	var name string

	if len(names) > 0 {
		name = names[0]
	} else {
		name = NewStackName()
	}

	requestBody := fmt.Sprintf(
		`{"name":"%s","description":"CF CLI integration test stack, please delete"}`,
		name,
	)
	session := CF("curl", "-v", "-X", "POST", "/v3/stacks", "-d", requestBody)

	Eventually(session).Should(Say("201 Created"))
	Eventually(session).Should(Exit(0))

	return name
}

// CreateStackWithGUID creates a stack with a random name and returns its name and guid
func CreateStackWithGUID() (string, string) {
	type StackStruct struct {
		GUID string `json:"guid"`
	}
	name := NewStackName()
	requestBody := fmt.Sprintf(
		`{"name":"%s","description":"CF CLI integration test stack, please delete"}`,
		name,
	)
	session := CF("curl", "-X", "POST", "/v3/stacks", "-d", requestBody)

	Eventually(session).Should(Exit(0))
	thisStack := StackStruct{}
	err := json.Unmarshal(session.Out.Contents(), &thisStack)
	Expect(err).ToNot(HaveOccurred())
	stackGUID := thisStack.GUID
	Expect(len(stackGUID)).ToNot(Equal(0))
	return name, stackGUID
}

// DeleteStack deletes a specific stack
func DeleteStack(name string) {
	session := CF("stack", "--guid", name)
	Eventually(session).Should(Exit(0))
	guid := strings.TrimSpace(string(session.Out.Contents()))

	session = CF("curl", "-v", "-X", "DELETE", "/v3/stacks/"+guid)

	Eventually(session).Should(Say("204 No Content"))
	Eventually(session).Should(Exit(0))
}

// EnsureMinimumNumberOfStacks ensures there are at least <num> stacks in the foundation by creating new ones if there
// are fewer than the specified number
func EnsureMinimumNumberOfStacks(num int) []string {
	var stacks []string
	for stacks = FetchStacks(); len(stacks) < num; {
		stacks = append(stacks, CreateStack())
	}
	return stacks
}
