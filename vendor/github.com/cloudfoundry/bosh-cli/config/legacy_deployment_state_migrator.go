package config

import (
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v2"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type LegacyDeploymentStateMigrator interface {
	MigrateIfExists(configPath string) (migrated bool, err error)
}

type legacyDeploymentStateMigrator struct {
	deploymentStateService DeploymentStateService
	fs                     boshsys.FileSystem
	uuidGenerator          boshuuid.Generator
	logger                 boshlog.Logger
	logTag                 string
}

func NewLegacyDeploymentStateMigrator(
	deploymentStateService DeploymentStateService,
	fs boshsys.FileSystem,
	uuidGenerator boshuuid.Generator,
	logger boshlog.Logger,
) LegacyDeploymentStateMigrator {
	return &legacyDeploymentStateMigrator{
		deploymentStateService: deploymentStateService,
		fs:            fs,
		uuidGenerator: uuidGenerator,
		logger:        logger,
		logTag:        "legacyDeploymentStateMigrator",
	}
}

func (m *legacyDeploymentStateMigrator) MigrateIfExists(configPath string) (migrated bool, err error) {
	if !m.fs.FileExists(configPath) {
		return false, nil
	}

	deploymentState, err := m.migrate(configPath)
	if err != nil {
		return false, err
	}

	err = m.deploymentStateService.Save(deploymentState)
	if err != nil {
		return false, bosherr.WrapError(err, "Saving migrated deployment state")
	}

	err = m.fs.RemoveAll(configPath)
	if err != nil {
		return false, bosherr.WrapError(err, "Deleting legacy deployment state")
	}

	return true, nil
}

func (m *legacyDeploymentStateMigrator) migrate(configPath string) (deploymentState DeploymentState, err error) {
	m.logger.Info(m.logTag, "Migrating legacy bosh-deployments.yml")

	bytes, err := m.fs.ReadFile(configPath)
	if err != nil {
		return deploymentState, bosherr.WrapErrorf(err, "Reading legacy deployment state file '%s'", configPath)
	}

	// go-yaml does not currently support ':' as the first character in a key.
	regex := regexp.MustCompile("\n([- ]) :")
	parsableString := regex.ReplaceAllString(string(bytes), "\n$1 ")

	m.logger.Debug(m.logTag, "Processed legacy bosh-deployments.yml:\n%s", parsableString)

	var legacyDeploymentState legacyDeploymentState
	err = yaml.Unmarshal([]byte(parsableString), &legacyDeploymentState)
	if err != nil {
		return deploymentState, bosherr.WrapError(err, "Parsing job manifest")
	}

	m.logger.Debug(m.logTag, "Parsed legacy bosh-deployments.yml: %#v", legacyDeploymentState)

	if (len(legacyDeploymentState.Instances) > 0) && (legacyDeploymentState.Instances[0].UUID != "") {
		deploymentState.DirectorID = legacyDeploymentState.Instances[0].UUID
	} else {
		uuid, err := m.uuidGenerator.Generate()
		if err != nil {
			return deploymentState, bosherr.WrapError(err, "Generating UUID")
		}
		deploymentState.DirectorID = uuid
	}

	deploymentState.Disks = []DiskRecord{}
	deploymentState.Stemcells = []StemcellRecord{}
	deploymentState.Releases = []ReleaseRecord{}

	if len(legacyDeploymentState.Instances) > 0 {
		instance := legacyDeploymentState.Instances[0]
		diskCID := instance.DiskCID
		if diskCID != "" {
			uuid, err := m.uuidGenerator.Generate()
			if err != nil {
				return deploymentState, bosherr.WrapError(err, "Generating UUID")
			}

			deploymentState.CurrentDiskID = uuid
			deploymentState.Disks = []DiskRecord{
				{
					ID:              uuid,
					CID:             diskCID,
					Size:            0,
					CloudProperties: biproperty.Map{},
				},
			}
		}

		vmCID := instance.VMCID
		if vmCID != "" {
			deploymentState.CurrentVMCID = vmCID
		}

		stemcellCID := instance.StemcellCID
		if stemcellCID != "" {
			uuid, err := m.uuidGenerator.Generate()
			if err != nil {
				return deploymentState, bosherr.WrapError(err, "Generating UUID")
			}

			stemcellName := instance.StemcellName
			if stemcellName == "" {
				stemcellName = "unknown-stemcell"
			}

			deploymentState.Stemcells = []StemcellRecord{
				{
					ID:      uuid,
					Name:    stemcellName,
					Version: "", // unknown, will never match new stemcell
					CID:     stemcellCID,
				},
			}
		}
	}

	m.logger.Debug(m.logTag, "New deployment.json (migrated from legacy bosh-deployments.yml): %#v", deploymentState)

	return deploymentState, nil
}

type legacyDeploymentState struct {
	Instances []instance `yaml:"instances"`
}

type instance struct {
	UUID         string `yaml:"uuid"`
	VMCID        string `yaml:"vm_cid"`
	DiskCID      string `yaml:"disk_cid"`
	StemcellCID  string `yaml:"stemcell_cid"`
	StemcellName string `yaml:"stemcell_name"`
}

func LegacyDeploymentStatePath(deploymentManifestPath string) string {
	return filepath.Join(filepath.Dir(deploymentManifestPath), "bosh-deployments.yml")
}
