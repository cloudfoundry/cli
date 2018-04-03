package config

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type DeploymentState struct {
	DirectorID         string           `json:"director_id"`
	InstallationID     string           `json:"installation_id"`
	CurrentVMCID       string           `json:"current_vm_cid"`
	CurrentStemcellID  string           `json:"current_stemcell_id"`
	CurrentDiskID      string           `json:"current_disk_id"`
	CurrentReleaseIDs  []string         `json:"current_release_ids"`
	CurrentManifestSHA string           `json:"current_manifest_sha"`
	Disks              []DiskRecord     `json:"disks"`
	Stemcells          []StemcellRecord `json:"stemcells"`
	Releases           []ReleaseRecord  `json:"releases"`
}

type StemcellRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	CID     string `json:"cid"`
}

type DiskRecord struct {
	ID              string         `json:"id"`
	CID             string         `json:"cid"`
	Size            int            `json:"size"`
	CloudProperties biproperty.Map `json:"cloud_properties"`
}

type ReleaseRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DeploymentStateService interface {
	Path() string
	Exists() bool
	Load() (DeploymentState, error)
	Save(DeploymentState) error
	Cleanup() error
}
