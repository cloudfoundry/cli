package stemcellfakes

import (
	"fmt"

	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	biui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type FakeManager struct {
	UploadInputs   []UploadInput
	uploadBehavior map[UploadInput]uploadOutput

	findUnusedOutput findUnusedOutput

	DeleteUnusedCalledTimes int
	DeleteUnusedErr         error
}

type UploadInput struct {
	Stemcell bistemcell.ExtractedStemcell
	Stage    biui.Stage
}

type uploadOutput struct {
	stemcell bistemcell.CloudStemcell
	err      error
}

type findUnusedOutput struct {
	stemcells []bistemcell.CloudStemcell
	err       error
}

func NewFakeManager() *FakeManager {
	return &FakeManager{
		UploadInputs:   []UploadInput{},
		uploadBehavior: map[UploadInput]uploadOutput{},
	}
}

func (m *FakeManager) FindCurrent() ([]bistemcell.CloudStemcell, error) {
	return []bistemcell.CloudStemcell{}, bosherr.Error("FakeManager.FindCurrent() not implemented (yet)")
}

func (m *FakeManager) Upload(stemcell bistemcell.ExtractedStemcell, stage biui.Stage) (bistemcell.CloudStemcell, error) {
	input := UploadInput{
		Stemcell: stemcell,
		Stage:    stage,
	}
	m.UploadInputs = append(m.UploadInputs, input)
	output, found := m.uploadBehavior[input]
	if !found {
		return nil, fmt.Errorf("Unsupported Upload Input: %#v", stemcell)
	}

	return output.stemcell, output.err
}

func (m *FakeManager) FindUnused() ([]bistemcell.CloudStemcell, error) {
	return m.findUnusedOutput.stemcells, m.findUnusedOutput.err
}

func (m *FakeManager) DeleteUnused(eventLoggerStage biui.Stage) error {
	m.DeleteUnusedCalledTimes++
	return m.DeleteUnusedErr
}

func (m *FakeManager) SetUploadBehavior(
	extractedStemcell bistemcell.ExtractedStemcell,
	stage biui.Stage,
	cloudStemcell bistemcell.CloudStemcell,
	err error,
) {
	input := UploadInput{
		Stemcell: extractedStemcell,
		Stage:    stage,
	}
	m.uploadBehavior[input] = uploadOutput{stemcell: cloudStemcell, err: err}
}

func (m *FakeManager) SetFindUnusedBehavior(
	stemcells []bistemcell.CloudStemcell,
	err error,
) {
	m.findUnusedOutput = findUnusedOutput{
		stemcells: stemcells,
		err:       err,
	}
}
