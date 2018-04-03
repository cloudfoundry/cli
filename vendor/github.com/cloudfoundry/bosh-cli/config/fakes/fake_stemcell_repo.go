package fakes

import (
	"fmt"

	biconfig "github.com/cloudfoundry/bosh-cli/config"
	bitestutils "github.com/cloudfoundry/bosh-cli/testutils"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type StemcellRepoSaveInput struct {
	Name    string
	Version string
	CID     string
}

type StemcellRepoSaveOutput struct {
	stemcellRecord biconfig.StemcellRecord
	err            error
}

type StemcellRepoFindInput struct {
	Name    string
	Version string
}

type StemcellRepoFindOutput struct {
	stemcellRecord biconfig.StemcellRecord
	found          bool
	err            error
}

type FindCurrentOutput struct {
	stemcellRecord biconfig.StemcellRecord
	found          bool
	err            error
}

type FakeStemcellRepo struct {
	SaveBehavior map[string]StemcellRepoSaveOutput
	SaveInputs   []StemcellRepoSaveInput
	FindBehavior map[string]StemcellRepoFindOutput
	FindInputs   []StemcellRepoFindInput

	UpdateCurrentRecordID string
	UpdateCurrentErr      error

	findCurrentOutput FindCurrentOutput

	ClearCurrentCalled bool
	ClearCurrentErr    error

	DeleteStemcellRecords []biconfig.StemcellRecord
	DeleteErr             error

	AllStemcellRecords []biconfig.StemcellRecord
	AllErr             error
}

func NewFakeStemcellRepo() *FakeStemcellRepo {
	return &FakeStemcellRepo{
		FindBehavior: map[string]StemcellRepoFindOutput{},
		FindInputs:   []StemcellRepoFindInput{},
		SaveBehavior: map[string]StemcellRepoSaveOutput{},
		SaveInputs:   []StemcellRepoSaveInput{},
	}
}

func (fr *FakeStemcellRepo) UpdateCurrent(recordID string) error {
	fr.UpdateCurrentRecordID = recordID
	return fr.UpdateCurrentErr
}

func (fr *FakeStemcellRepo) FindCurrent() (biconfig.StemcellRecord, bool, error) {
	return fr.findCurrentOutput.stemcellRecord, fr.findCurrentOutput.found, fr.findCurrentOutput.err
}

func (fr *FakeStemcellRepo) ClearCurrent() error {
	fr.ClearCurrentCalled = true
	return fr.ClearCurrentErr
}

func (fr *FakeStemcellRepo) Delete(stemcellRecord biconfig.StemcellRecord) error {
	fr.DeleteStemcellRecords = append(fr.DeleteStemcellRecords, stemcellRecord)
	return fr.DeleteErr
}

func (fr *FakeStemcellRepo) All() ([]biconfig.StemcellRecord, error) {
	return fr.AllStemcellRecords, fr.AllErr
}

func (fr *FakeStemcellRepo) Save(name, version, cid string) (biconfig.StemcellRecord, error) {
	input := StemcellRepoSaveInput{
		Name:    name,
		Version: version,
		CID:     cid,
	}
	fr.SaveInputs = append(fr.SaveInputs, input)

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		return biconfig.StemcellRecord{}, bosherr.WrapError(marshalErr, "Marshaling Save input")
	}

	output, found := fr.SaveBehavior[inputString]
	if !found {
		return biconfig.StemcellRecord{}, fmt.Errorf("Unsupported Save Input: %s", inputString)
	}

	return output.stemcellRecord, output.err
}

func (fr *FakeStemcellRepo) SetSaveBehavior(name, version, cid string, stemcellRecord biconfig.StemcellRecord, err error) error {
	input := StemcellRepoSaveInput{
		Name:    name,
		Version: version,
		CID:     cid,
	}

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Save input")
	}

	fr.SaveBehavior[inputString] = StemcellRepoSaveOutput{
		stemcellRecord: stemcellRecord,
		err:            err,
	}

	return nil
}

func (fr *FakeStemcellRepo) Find(name, version string) (biconfig.StemcellRecord, bool, error) {
	input := StemcellRepoFindInput{
		Name:    name,
		Version: version,
	}
	fr.FindInputs = append(fr.FindInputs, input)

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		return biconfig.StemcellRecord{}, false, bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	output, found := fr.FindBehavior[inputString]
	if !found {
		return biconfig.StemcellRecord{}, false, fmt.Errorf("Unsupported Find Input: %s", inputString)
	}

	return output.stemcellRecord, output.found, output.err
}

func (fr *FakeStemcellRepo) SetFindBehavior(name, version string, foundRecord biconfig.StemcellRecord, found bool, err error) error {
	input := StemcellRepoFindInput{
		Name:    name,
		Version: version,
	}

	inputString, marshalErr := bitestutils.MarshalToString(input)
	if marshalErr != nil {
		return bosherr.WrapError(marshalErr, "Marshaling Find input")
	}

	fr.FindBehavior[inputString] = StemcellRepoFindOutput{
		stemcellRecord: foundRecord,
		found:          found,
		err:            err,
	}

	return nil
}

func (fr *FakeStemcellRepo) SetFindCurrentBehavior(foundRecord biconfig.StemcellRecord, found bool, err error) error {
	fr.findCurrentOutput = FindCurrentOutput{
		stemcellRecord: foundRecord,
		found:          found,
		err:            err,
	}

	return nil
}
