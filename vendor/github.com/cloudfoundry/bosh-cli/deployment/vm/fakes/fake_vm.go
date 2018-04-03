package fakes

import (
	"time"

	biagentclient "github.com/cloudfoundry/bosh-agent/agentclient"
	bias "github.com/cloudfoundry/bosh-agent/agentclient/applyspec"
	bidisk "github.com/cloudfoundry/bosh-cli/deployment/disk"
	bideplmanifest "github.com/cloudfoundry/bosh-cli/deployment/manifest"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type FakeVM struct {
	cid string

	ExistsCalled int
	ExistsFound  bool
	ExistsErr    error

	AgentClientReturn biagentclient.AgentClient

	UpdateDisksInputs []UpdateDisksInput
	UpdateDisksDisks  []bidisk.Disk
	UpdateDisksErr    error

	ApplyInputs []ApplyInput
	ApplyErr    error

	StartCalled int
	StartErr    error

	AttachDiskInputs   []AttachDiskInput
	attachDiskBehavior map[string]error

	DetachDiskInputs   []DetachDiskInput
	detachDiskBehavior map[string]error

	WaitUntilReadyInputs []WaitUntilReadyInput
	WaitUntilReadyErr    error

	WaitToBeRunningInputs []WaitInput
	WaitToBeRunningErr    error

	DeleteCalled int
	DeleteErr    error

	StopCalled int
	StopErr    error

	ListDisksDisks []bidisk.Disk
	ListDisksErr   error

	UnmountDiskInputs []UnmountDiskInput
	UnmountDiskErr    error

	MigrateDiskCalledTimes int
	MigrateDiskErr         error

	RunScriptInputs []string
	RunScriptErrors map[string]error

	GetStateResult biagentclient.AgentState
	GetStateCalled int
	GetStateErr    error
}

type UpdateDisksInput struct {
	DiskPool bideplmanifest.DiskPool
	Stage    biui.Stage
}

type ApplyInput struct {
	ApplySpec bias.ApplySpec
}

type WaitUntilReadyInput struct {
	Timeout time.Duration
	Delay   time.Duration
}

type WaitInput struct {
	MaxAttempts int
	Delay       time.Duration
}

type AttachDiskInput struct {
	Disk bidisk.Disk
}

type DetachDiskInput struct {
	Disk bidisk.Disk
}

type UnmountDiskInput struct {
	Disk bidisk.Disk
}

func NewFakeVM(cid string) *FakeVM {
	return &FakeVM{
		ExistsFound:           true,
		ApplyInputs:           []ApplyInput{},
		WaitUntilReadyInputs:  []WaitUntilReadyInput{},
		WaitToBeRunningInputs: []WaitInput{},
		AttachDiskInputs:      []AttachDiskInput{},
		DetachDiskInputs:      []DetachDiskInput{},
		UnmountDiskInputs:     []UnmountDiskInput{},
		attachDiskBehavior:    map[string]error{},
		detachDiskBehavior:    map[string]error{},
		cid:                   cid,
		RunScriptErrors:       map[string]error{},
	}
}

func (vm *FakeVM) CID() string {
	return vm.cid
}

func (vm *FakeVM) Exists() (bool, error) {
	vm.ExistsCalled++
	return vm.ExistsFound, vm.ExistsErr
}

func (vm *FakeVM) AgentClient() biagentclient.AgentClient {
	return vm.AgentClientReturn
}

func (vm *FakeVM) WaitUntilReady(timeout time.Duration, delay time.Duration) error {
	vm.WaitUntilReadyInputs = append(vm.WaitUntilReadyInputs, WaitUntilReadyInput{
		Timeout: timeout,
		Delay:   delay,
	})
	return vm.WaitUntilReadyErr
}

func (vm *FakeVM) UpdateDisks(diskPool bideplmanifest.DiskPool, eventLoggerStage biui.Stage) ([]bidisk.Disk, error) {
	vm.UpdateDisksInputs = append(vm.UpdateDisksInputs, UpdateDisksInput{
		DiskPool: diskPool,
		Stage:    eventLoggerStage,
	})
	return vm.UpdateDisksDisks, vm.UpdateDisksErr
}

func (vm *FakeVM) Apply(applySpec bias.ApplySpec) error {
	vm.ApplyInputs = append(vm.ApplyInputs, ApplyInput{
		ApplySpec: applySpec,
	})

	return vm.ApplyErr
}

func (vm *FakeVM) Start() error {
	vm.StartCalled++
	return vm.StartErr
}

func (vm *FakeVM) WaitToBeRunning(maxAttempts int, delay time.Duration) error {
	vm.WaitToBeRunningInputs = append(vm.WaitToBeRunningInputs, WaitInput{
		MaxAttempts: maxAttempts,
		Delay:       delay,
	})
	return vm.WaitToBeRunningErr
}

func (vm *FakeVM) AttachDisk(disk bidisk.Disk) error {
	vm.AttachDiskInputs = append(vm.AttachDiskInputs, AttachDiskInput{
		Disk: disk,
	})

	return vm.attachDiskBehavior[disk.CID()]
}

func (vm *FakeVM) DetachDisk(disk bidisk.Disk) error {
	vm.DetachDiskInputs = append(vm.DetachDiskInputs, DetachDiskInput{
		Disk: disk,
	})

	return vm.detachDiskBehavior[disk.CID()]
}

func (vm *FakeVM) UnmountDisk(disk bidisk.Disk) error {
	vm.UnmountDiskInputs = append(vm.UnmountDiskInputs, UnmountDiskInput{
		Disk: disk,
	})

	return vm.UnmountDiskErr
}

func (vm *FakeVM) MigrateDisk() error {
	vm.MigrateDiskCalledTimes++

	return vm.MigrateDiskErr
}

func (vm *FakeVM) Stop() error {
	vm.StopCalled++
	return vm.StopErr
}

func (vm *FakeVM) Disks() ([]bidisk.Disk, error) {
	return vm.ListDisksDisks, vm.ListDisksErr
}

func (vm *FakeVM) RunScript(script string, options map[string]interface{}) error {
	vm.RunScriptInputs = append(vm.RunScriptInputs, script)
	return vm.RunScriptErrors[script]
}

func (vm *FakeVM) Delete() error {
	vm.DeleteCalled++
	return vm.DeleteErr
}

func (vm *FakeVM) SetAttachDiskBehavior(disk bidisk.Disk, err error) {
	vm.attachDiskBehavior[disk.CID()] = err
}

func (vm *FakeVM) SetDetachDiskBehavior(disk bidisk.Disk, err error) {
	vm.detachDiskBehavior[disk.CID()] = err
}

func (vm *FakeVM) GetState() (biagentclient.AgentState, error) {
	vm.GetStateCalled++
	return vm.GetStateResult, vm.GetStateErr
}
