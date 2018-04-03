package testutils

import (
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

func GenerateDeploymentManifest(deploymentManifestFilePath string, fs boshsys.FileSystem, manifestContents string) error {
	return fs.WriteFileString(deploymentManifestFilePath, manifestContents)
}
