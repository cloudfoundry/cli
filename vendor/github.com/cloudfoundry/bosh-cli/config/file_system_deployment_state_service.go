package config

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type fileSystemDeploymentStateService struct {
	configPath    string
	fs            boshsys.FileSystem
	uuidGenerator boshuuid.Generator
	logger        boshlog.Logger
	logTag        string
}

func NewFileSystemDeploymentStateService(fs boshsys.FileSystem, uuidGenerator boshuuid.Generator, logger boshlog.Logger, deploymentStatePath string) DeploymentStateService {
	return &fileSystemDeploymentStateService{
		configPath:    deploymentStatePath,
		fs:            fs,
		uuidGenerator: uuidGenerator,
		logger:        logger,
		logTag:        "config",
	}
}

func DeploymentStatePath(deploymentManifestPath string, deploymentStatePath string) string {
	if deploymentStatePath != "" {
		return deploymentStatePath
	}

	baseFileName := filepath.Base(strings.TrimSuffix(deploymentManifestPath, filepath.Ext(deploymentManifestPath)))
	return filepath.Join(filepath.Dir(deploymentManifestPath), fmt.Sprintf("%s-state.json", baseFileName))
}

func (s *fileSystemDeploymentStateService) Path() string {
	return s.configPath
}

func (s *fileSystemDeploymentStateService) Exists() bool {
	return s.fs.FileExists(s.configPath)
}

func (s *fileSystemDeploymentStateService) Load() (DeploymentState, error) {
	if s.configPath == "" {
		panic("configPath not yet set!")
	}

	s.logger.Debug(s.logTag, "Loading deployment state: %s", s.configPath)

	deploymentState := &DeploymentState{}

	if s.fs.FileExists(s.configPath) {
		deploymentStateFileContents, err := s.fs.ReadFile(s.configPath)
		if err != nil {
			return DeploymentState{}, bosherr.WrapErrorf(err, "Reading deployment state file '%s'", s.configPath)
		}
		s.logger.Debug(s.logTag, "Deployment File Contents %#s", deploymentStateFileContents)

		err = json.Unmarshal(deploymentStateFileContents, deploymentState)
		if err != nil {
			return DeploymentState{}, bosherr.WrapErrorf(err, "Unmarshalling deployment state file '%s'", s.configPath)
		}
	}

	err := s.initDefaults(deploymentState)
	if err != nil {
		return DeploymentState{}, bosherr.WrapErrorf(err, "Initializing deployment state defaults")
	}

	return *deploymentState, nil
}

func (s *fileSystemDeploymentStateService) Save(deploymentState DeploymentState) error {
	if s.configPath == "" {
		panic("configPath not yet set!")
	}

	s.logger.Debug(s.logTag, "Saving deployment state %#v", deploymentState)

	jsonContent, err := json.MarshalIndent(deploymentState, "", "    ")
	if err != nil {
		return bosherr.WrapError(err, "Marshalling deployment state into JSON")
	}

	err = s.fs.WriteFile(s.configPath, jsonContent)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing deployment state file '%s'", s.configPath)
	}

	return nil
}

func (s *fileSystemDeploymentStateService) initDefaults(deploymentState *DeploymentState) error {
	if deploymentState.DirectorID == "" {
		uuid, err := s.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating DirectorID")
		}
		deploymentState.DirectorID = uuid

		err = s.Save(*deploymentState)
		if err != nil {
			return bosherr.WrapError(err, "Saving deployment state")
		}
	}

	return nil
}

func (s *fileSystemDeploymentStateService) Cleanup() error {
	err := s.fs.RemoveAll(s.configPath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Could not delete deployment state file %s", s.configPath)
	}
	return nil
}
