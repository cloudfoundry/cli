package uaa

import (
	"sort"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Prompt struct {
	Key   string // e.g. "username"
	Type  string // e.g. "text", "password"
	Label string // e.g. "Username"
}

type PromptAnswer struct {
	Key   string // e.g. "username"
	Value string
}

func (p Prompt) IsPassword() bool { return p.Type == "password" }

type PromptsResp struct {
	Prompts map[string][]string // e.g. {"username": ["text", "Username"]}
}

func (u UAAImpl) Prompts() ([]Prompt, error) {
	resp, err := u.client.Prompts()
	if err != nil {
		return nil, err
	}

	var prompts []Prompt

	for key, pair := range resp.Prompts {
		prompts = append(prompts, Prompt{
			Key:   key,
			Type:  pair[0],
			Label: pair[1],
		})
	}

	// Ideally UAA would sort prompts...
	sort.Sort(PromptSorting(prompts))

	return prompts, nil
}

func (c Client) Prompts() (PromptsResp, error) {
	var resp PromptsResp

	err := c.clientRequest.Get("/login", &resp)
	if err != nil {
		return resp, bosherr.WrapError(err, "Requesting UAA prompts")
	}

	return resp, nil
}

type PromptSorting []Prompt

func (s PromptSorting) Len() int           { return len(s) }
func (s PromptSorting) Less(i, j int) bool { return s[i].Type > s[j].Type }
func (s PromptSorting) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
