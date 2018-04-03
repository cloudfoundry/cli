package stemcellfakes

import (
	"fmt"

	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
)

type ExtractInput struct {
	TarballPath string
}

type extractOutput struct {
	stemcell bistemcell.ExtractedStemcell
	err      error
}

type FakeExtractor struct {
	ExtractInputs   []ExtractInput
	extractBehavior map[ExtractInput]extractOutput
}

func NewFakeExtractor() *FakeExtractor {
	return &FakeExtractor{
		ExtractInputs:   []ExtractInput{},
		extractBehavior: map[ExtractInput]extractOutput{},
	}
}

func (e *FakeExtractor) Extract(tarballPath string) (bistemcell.ExtractedStemcell, error) {
	input := ExtractInput{
		TarballPath: tarballPath,
	}
	e.ExtractInputs = append(e.ExtractInputs, input)
	output, found := e.extractBehavior[input]
	if !found {
		return nil, fmt.Errorf("Unsupported Upload Input: %s", tarballPath)
	}

	return output.stemcell, output.err
}

func (e *FakeExtractor) SetExtractBehavior(
	tarballPath string,
	extractedStemcell bistemcell.ExtractedStemcell,
	err error,
) {
	input := ExtractInput{
		TarballPath: tarballPath,
	}
	e.extractBehavior[input] = extractOutput{stemcell: extractedStemcell, err: err}
}
