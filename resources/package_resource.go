package resources

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
)

// Package represents a Cloud Controller V3 Package.
type Package struct {
	// CreatedAt is the time with zone when the object was created.
	CreatedAt string

	// DockerImage is the registry address of the docker image.
	DockerImage string

	// DockerPassword is the password for the docker image's registry.
	DockerPassword string

	// DockerUsername is the username for the docker image's registry.
	DockerUsername string

	// GUID is the unique identifier of the package.
	GUID string

	// Links are links to related resources.
	Links APILinks

	// Relationships are a list of relationships to other resources.
	Relationships Relationships

	// State is the state of the package.
	State constant.PackageState

	// Type is the package type.
	Type constant.PackageType
}

// MarshalJSON converts a Package into a Cloud Controller Package.
func (p Package) MarshalJSON() ([]byte, error) {
	type ccPackageData struct {
		Image    string `json:"image,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
	var ccPackage struct {
		GUID          string                `json:"guid,omitempty"`
		CreatedAt     string                `json:"created_at,omitempty"`
		Links         APILinks              `json:"links,omitempty"`
		Relationships Relationships         `json:"relationships,omitempty"`
		State         constant.PackageState `json:"state,omitempty"`
		Type          constant.PackageType  `json:"type,omitempty"`
		Data          *ccPackageData        `json:"data,omitempty"`
	}

	ccPackage.GUID = p.GUID
	ccPackage.CreatedAt = p.CreatedAt
	ccPackage.Links = p.Links
	ccPackage.Relationships = p.Relationships
	ccPackage.State = p.State
	ccPackage.Type = p.Type
	if p.DockerImage != "" {
		ccPackage.Data = &ccPackageData{
			Image:    p.DockerImage,
			Username: p.DockerUsername,
			Password: p.DockerPassword,
		}
	}

	return json.Marshal(ccPackage)
}

// UnmarshalJSON helps unmarshal a Cloud Controller Package response.
func (p *Package) UnmarshalJSON(data []byte) error {
	var ccPackage struct {
		GUID          string                `json:"guid,omitempty"`
		CreatedAt     string                `json:"created_at,omitempty"`
		Links         APILinks              `json:"links,omitempty"`
		Relationships Relationships         `json:"relationships,omitempty"`
		State         constant.PackageState `json:"state,omitempty"`
		Type          constant.PackageType  `json:"type,omitempty"`
		Data          struct {
			Image    string `json:"image"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"data"`
	}
	err := cloudcontroller.DecodeJSON(data, &ccPackage)
	if err != nil {
		return err
	}

	p.GUID = ccPackage.GUID
	p.CreatedAt = ccPackage.CreatedAt
	p.Links = ccPackage.Links
	p.Relationships = ccPackage.Relationships
	p.State = ccPackage.State
	p.Type = ccPackage.Type
	p.DockerImage = ccPackage.Data.Image
	p.DockerUsername = ccPackage.Data.Username
	p.DockerPassword = ccPackage.Data.Password

	return nil
}
