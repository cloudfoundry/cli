package fakes

import (
	"github.com/cloudfoundry/bosh-cli/cloud"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type FakeCloud struct {
	CreateStemcellInputs []CreateStemcellInput
	CreateStemcellCID    string
	CreateStemcellErr    error

	HasVMInput HasVMInput
	HasVMFound bool
	HasVMErr   error

	CreateVMInput CreateVMInput
	CreateVMCID   string
	CreateVMErr   error

	CreateDiskInput CreateDiskInput
	CreateDiskCID   string
	CreateDiskErr   error

	AttachDiskInput AttachDiskInput
	AttachDiskErr   error

	DetachDiskInput DetachDiskInput
	DetachDiskErr   error

	DeleteVMInput DeleteVMInput
	DeleteVMErr   error

	DeleteDiskInputs []DeleteDiskInput
	DeleteDiskErr    error

	DeleteStemcellInputs []DeleteStemcellInput
	DeleteStemcellErr    error

	SetVMMetadataCid      string
	SetVMMetadataMetadata cloud.VMMetadata
	SetVMMetadataError    error

	SetDiskMetadataCid      string
	SetDiskMetadataMetadata cloud.DiskMetadata
	SetDiskMetadataError    error
}

type CreateStemcellInput struct {
	ImagePath       string
	CloudProperties biproperty.Map
}

type HasVMInput struct {
	VMCID string
}

type CreateVMInput struct {
	AgentID            string
	StemcellCID        string
	CloudProperties    biproperty.Map
	NetworksInterfaces map[string]biproperty.Map
	Env                biproperty.Map
}

type CreateDiskInput struct {
	Size            int
	CloudProperties biproperty.Map
	InstanceID      string
}

type AttachDiskInput struct {
	VMCID   string
	DiskCID string
}

type DetachDiskInput struct {
	VMCID   string
	DiskCID string
}

type DeleteVMInput struct {
	VMCID string
}

type DeleteDiskInput struct {
	DiskCID string
}

type DeleteStemcellInput struct {
	StemcellCID string
}

func NewFakeCloud() *FakeCloud {
	return &FakeCloud{
		CreateStemcellInputs: []CreateStemcellInput{},
		DeleteDiskInputs:     []DeleteDiskInput{},
	}
}

func (c *FakeCloud) CreateStemcell(imagePath string, cloudProperties biproperty.Map) (string, error) {
	c.CreateStemcellInputs = append(c.CreateStemcellInputs, CreateStemcellInput{
		ImagePath:       imagePath,
		CloudProperties: cloudProperties,
	})

	return c.CreateStemcellCID, c.CreateStemcellErr
}

func (c *FakeCloud) DeleteStemcell(stemcellCID string) error {
	c.DeleteStemcellInputs = append(c.DeleteStemcellInputs, DeleteStemcellInput{
		StemcellCID: stemcellCID,
	})
	return c.DeleteStemcellErr
}

func (c *FakeCloud) HasVM(vmCID string) (bool, error) {
	c.HasVMInput = HasVMInput{
		VMCID: vmCID,
	}
	return c.HasVMFound, c.HasVMErr
}

func (c *FakeCloud) CreateVM(
	agentID string,
	stemcellCID string,
	cloudProperties biproperty.Map,
	networksInterfaces map[string]biproperty.Map,
	env biproperty.Map,
) (string, error) {
	c.CreateVMInput = CreateVMInput{
		AgentID:            agentID,
		StemcellCID:        stemcellCID,
		CloudProperties:    cloudProperties,
		NetworksInterfaces: networksInterfaces,
		Env:                env,
	}

	return c.CreateVMCID, c.CreateVMErr
}

func (c *FakeCloud) SetVMMetadata(cid string, metadata cloud.VMMetadata) error {
	c.SetVMMetadataCid = cid
	c.SetVMMetadataMetadata = metadata
	return c.SetVMMetadataError
}

func (c *FakeCloud) SetDiskMetadata(cid string, metadata cloud.DiskMetadata) error {
	c.SetDiskMetadataCid = cid
	c.SetDiskMetadataMetadata = metadata
	return c.SetDiskMetadataError
}

func (c *FakeCloud) CreateDisk(
	size int,
	cloudProperties biproperty.Map,
	instanceID string,
) (string, error) {
	c.CreateDiskInput = CreateDiskInput{
		Size:            size,
		CloudProperties: cloudProperties,
		InstanceID:      instanceID,
	}

	return c.CreateDiskCID, c.CreateDiskErr
}

func (c *FakeCloud) AttachDisk(vmCID, diskCID string) error {
	c.AttachDiskInput = AttachDiskInput{
		VMCID:   vmCID,
		DiskCID: diskCID,
	}
	return c.AttachDiskErr
}

func (c *FakeCloud) DetachDisk(vmCID, diskCID string) error {
	c.DetachDiskInput = DetachDiskInput{
		VMCID:   vmCID,
		DiskCID: diskCID,
	}
	return c.DetachDiskErr
}

func (c *FakeCloud) DeleteVM(vmCID string) error {
	c.DeleteVMInput = DeleteVMInput{
		VMCID: vmCID,
	}
	return c.DeleteVMErr
}

func (c *FakeCloud) DeleteDisk(diskCID string) error {
	c.DeleteDiskInputs = append(c.DeleteDiskInputs, DeleteDiskInput{
		DiskCID: diskCID,
	})
	return c.DeleteDiskErr
}

func (c *FakeCloud) String() string {
	return "FakeCloud{}"
}
