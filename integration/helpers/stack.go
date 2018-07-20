package helpers

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ccStacks struct {
	Resources []struct {
		Entity struct {
			Name string `json:"name"`
		} `json:"entity"`
	} `json:"resources"`
}

func SkipIfOneStack() {
	if len(FetchStacks()) < 2 {
		Skip("test requires at least two stacks")
	}
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
