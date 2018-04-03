package fakes

import (
	"fmt"
	"strings"
)

type FakeDigestCalculator struct {
	calculateInputs       map[string]CalculateInput
	CalculateStringInputs map[string]string
}

func NewFakeDigestCalculator() *FakeDigestCalculator {
	return &FakeDigestCalculator{}
}

type CalculateInput struct {
	DigestStr string
	Err       error
}

func (c *FakeDigestCalculator) Calculate(path string) (string, error) {
	calculateInput := c.calculateInputs[path]
	return calculateInput.DigestStr, calculateInput.Err
}

func (c *FakeDigestCalculator) CalculateString(data string) string {
	if sha1, found := c.CalculateStringInputs[data]; found {
		return sha1
	}

	var availableData []string
	for key, _ := range c.CalculateStringInputs {
		availableData = append(availableData, key)
	}

	panic(fmt.Sprintf("Did not find SHA1 result for '%s'. Available result keys:'%s'", data, strings.Join(availableData, ", ")))
}

func (c *FakeDigestCalculator) SetCalculateBehavior(calculateInputs map[string]CalculateInput) {
	c.calculateInputs = calculateInputs
}
