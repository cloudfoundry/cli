package director

import (
	"io"
	"os"
	"time"

	"github.com/cloudfoundry/bosh-cli/ui"
	semver "github.com/cppforlife/go-semi-semantic/version"
)

//go:generate counterfeiter . Director

type Director interface {
	IsAuthenticated() (bool, error)
	WithContext(id string) Director
	Info() (Info, error)

	Locks() ([]Lock, error)

	CurrentTasks(TasksFilter) ([]Task, error)
	RecentTasks(int, TasksFilter) ([]Task, error)
	FindTask(int) (Task, error)
	FindTasksByContextId(string) ([]Task, error)

	Events(EventsFilter) ([]Event, error)
	Event(string) (Event, error)

	Deployments() ([]Deployment, error)
	FindDeployment(string) (Deployment, error)

	Releases() ([]Release, error)
	HasRelease(name, version string, stemcell OSVersionSlug) (bool, error)
	FindRelease(ReleaseSlug) (Release, error)
	FindReleaseSeries(ReleaseSeriesSlug) (ReleaseSeries, error)
	UploadReleaseURL(url, sha1 string, rebase, fix bool) error
	UploadReleaseFile(file UploadFile, rebase, fix bool) error
	MatchPackages(manifest interface{}, compiled bool) ([]string, error)

	Stemcells() ([]Stemcell, error)
	HasStemcell(name, version string) (bool, error)
	StemcellNeedsUpload(StemcellInfo) (bool, bool, error)
	FindStemcell(StemcellSlug) (Stemcell, error)
	UploadStemcellURL(url, sha1 string, fix bool) error
	UploadStemcellFile(file UploadFile, fix bool) error

	LatestConfig(configType string, name string) (Config, error)
	LatestConfigByID(configID string) (Config, error)
	ListConfigs(limit int, filter ConfigsFilter) ([]Config, error)
	UpdateConfig(configType string, name string, content []byte) (Config, error)
	DeleteConfig(configType string, name string) (bool, error)
	DeleteConfigByID(configID string) (bool, error)
	DiffConfig(configType string, name string, manifest []byte) (ConfigDiff, error)
	DiffConfigByID(fromID string, toID string) (ConfigDiff, error)

	LatestCloudConfig() (CloudConfig, error)
	UpdateCloudConfig([]byte) error
	DiffCloudConfig(manifest []byte) (ConfigDiff, error)

	LatestCPIConfig() (CPIConfig, error)
	UpdateCPIConfig([]byte) error
	DiffCPIConfig(manifest []byte, noRedact bool) (ConfigDiff, error)

	LatestRuntimeConfig(name string) (RuntimeConfig, error)
	UpdateRuntimeConfig(name string, manifest []byte) error
	DiffRuntimeConfig(name string, manifest []byte, noRedact bool) (ConfigDiff, error)

	FindOrphanDisk(string) (OrphanDisk, error)
	OrphanDisks() ([]OrphanDisk, error)
	OrphanDisk(string) error

	EnableResurrection(bool) error
	CleanUp(bool) error
	DownloadResourceUnchecked(blobstoreID string, out io.Writer) error
}

var _ Director = &DirectorImpl{}

type UploadFile interface {
	io.ReadCloser
	Stat() (os.FileInfo, error)
}

//go:generate counterfeiter . ReleaseArchive

type ReleaseArchive interface {
	Info() (string, string, error)
	File() (UploadFile, error)
}

//go:generate counterfeiter . StemcellArchive

type StemcellArchive interface {
	Info() (string, string, error)
	File() (UploadFile, error)
}

//go:generate counterfeiter . FileReporter

type FileReporter interface {
	TrackUpload(int64, io.ReadCloser) ui.ReadSeekCloser
	TrackDownload(int64, io.Writer) io.Writer
}

//go:generate counterfeiter . Deployment

type Deployment interface {
	Name() string
	Manifest() (string, error)
	CloudConfig() (string, error)
	Diff([]byte, bool) (DeploymentDiff, error)

	Releases() ([]Release, error)
	ExportRelease(ReleaseSlug, OSVersionSlug, []string) (ExportReleaseResult, error)

	Teams() ([]string, error)

	Stemcells() ([]Stemcell, error)
	VMInfos() ([]VMInfo, error)
	Instances() ([]Instance, error)
	InstanceInfos() ([]VMInfo, error)

	Errands() ([]Errand, error)
	RunErrand(string, bool, bool, []InstanceGroupOrInstanceSlug) ([]ErrandResult, error)

	ScanForProblems() ([]Problem, error)
	ResolveProblems([]ProblemAnswer) error

	Snapshots() ([]Snapshot, error)
	TakeSnapshots() error
	DeleteSnapshot(string) error
	DeleteSnapshots() error
	DeleteVM(string) error

	Variables() ([]VariableResult, error)

	// Deployment, pool or instance specifics
	Start(slug AllOrInstanceGroupOrInstanceSlug, opts StartOpts) error
	Stop(slug AllOrInstanceGroupOrInstanceSlug, opts StopOpts) error
	Restart(slug AllOrInstanceGroupOrInstanceSlug, opts RestartOpts) error
	Recreate(slug AllOrInstanceGroupOrInstanceSlug, opts RecreateOpts) error

	SetUpSSH(AllOrInstanceGroupOrInstanceSlug, SSHOpts) (SSHResult, error)
	CleanUpSSH(AllOrInstanceGroupOrInstanceSlug, SSHOpts) error

	// Instance specifics
	FetchLogs(AllOrInstanceGroupOrInstanceSlug, []string, bool) (LogsResult, error)
	TakeSnapshot(InstanceSlug) error
	Ignore(InstanceSlug, bool) error
	EnableResurrection(InstanceSlug, bool) error

	Update(manifest []byte, opts UpdateOpts) error
	Delete(force bool) error

	AttachDisk(slug InstanceSlug, diskCID string) error
}

type StartOpts struct {
	Canaries    string
	MaxInFlight string
}

type StopOpts struct {
	Canaries    string
	MaxInFlight string
	Force       bool
	SkipDrain   bool
	Hard        bool
}

type RestartOpts struct {
	Canaries    string
	MaxInFlight string
	Force       bool
	SkipDrain   bool
}

type RecreateOpts struct {
	Canaries    string
	MaxInFlight string
	Force       bool
	Fix         bool
	SkipDrain   bool
	DryRun      bool
}

type UpdateOpts struct {
	Recreate    bool
	Fix         bool
	SkipDrain   SkipDrains
	Canaries    string
	MaxInFlight string
	DryRun      bool
	Diff        DeploymentDiff
}

//go:generate counterfeiter . ReleaseSeries

type ReleaseSeries interface {
	Name() string
	Delete(force bool) error
}

//go:generate counterfeiter . Release

type Release interface {
	Name() string
	Version() semver.Version
	VersionMark(mark string) string
	CommitHashWithMark(mark string) string

	Jobs() ([]Job, error)
	Packages() ([]Package, error)

	Delete(force bool) error
}

//go:generate counterfeiter . Stemcell

type Stemcell interface {
	Name() string
	Version() semver.Version
	VersionMark(mark string) string

	OSName() string

	CPI() string
	CID() string

	Delete(force bool) error
}

type TasksFilter struct {
	All        bool
	Deployment string
}

type Task interface {
	ID() int
	StartedAt() time.Time
	LastActivityAt() time.Time

	State() string
	IsError() bool
	User() string
	DeploymentName() string
	ContextID() string

	Description() string
	Result() string

	EventOutput(TaskReporter) error
	CPIOutput(TaskReporter) error
	DebugOutput(TaskReporter) error
	ResultOutput(TaskReporter) error

	Cancel() error
}

//go:generate counterfeiter . TaskReporter

type TaskReporter interface {
	TaskStarted(int)
	TaskFinished(int, string)
	TaskOutputChunk(int, []byte)
}

//go:generate counterfeiter . OrphanDisk

type OrphanDisk interface {
	CID() string
	Size() uint64

	Deployment() Deployment
	InstanceName() string
	AZName() string

	OrphanedAt() time.Time

	Delete() error
}

type EventsFilter struct {
	BeforeID   string
	Before     string
	After      string
	Deployment string
	Task       string
	Instance   string
	User       string
	Action     string
	ObjectType string
	ObjectName string
}

//go:generate counterfeiter . Event

type Event interface {
	ID() string
	ParentID() string
	Timestamp() time.Time
	User() string
	Action() string
	ObjectType() string
	ObjectName() string
	TaskID() string
	DeploymentName() string
	Instance() string
	Context() map[string]interface{}
	Error() string
}
