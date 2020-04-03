package ccv3

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

//go:generate counterfeiter io.Reader

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

// CreatePackage creates a package with the given settings, Type and the
// ApplicationRelationship must be set.
func (client *Client) CreatePackage(pkg Package) (Package, Warnings, error) {
	var responseBody Package

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.PostPackageRequest,
		RequestBody:  pkg,
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetPackage returns the package with the given GUID.
func (client *Client) GetPackage(packageGUID string) (Package, Warnings, error) {
	var responseBody Package

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName:  internal.GetPackageRequest,
		URIParams:    internal.Params{"package_guid": packageGUID},
		ResponseBody: &responseBody,
	})

	return responseBody, warnings, err
}

// GetPackages returns the list of packages.
func (client *Client) GetPackages(query ...Query) ([]Package, Warnings, error) {
	var resources []Package

	_, warnings, err := client.MakeListRequest(RequestParams{
		RequestName:  internal.GetPackagesRequest,
		Query:        query,
		ResponseBody: Package{},
		AppendToList: func(item interface{}) error {
			resources = append(resources, item.(Package))
			return nil
		},
	})

	return resources, warnings, err
}

// UploadBitsPackage uploads the newResources and a list of existing resources
// to the cloud controller. An updated package is returned. The function will
// act differently given the following Readers:
//   - io.ReadSeeker: Will function properly on retry.
//   - io.Reader: Will return a ccerror.PipeSeekError on retry.
//   - nil: Will not add the "application" section to the request. The newResourcesLength is ignored in this case.
//
// Note: In order to determine if package creation is successful, poll the
// Package's state field for more information.
func (client *Client) UploadBitsPackage(pkg Package, matchedResources []Resource, newResources io.Reader, newResourcesLength int64) (Package, Warnings, error) {
	if matchedResources == nil {
		return Package{}, nil, ccerror.NilObjectError{Object: "matchedResources"}
	}

	if newResources == nil {
		return client.uploadExistingResourcesOnly(pkg.GUID, matchedResources)
	}

	return client.uploadNewAndExistingResources(pkg.GUID, matchedResources, newResources, newResourcesLength)
}

// UploadPackage uploads a file to a given package's Upload resource. Note:
// fileToUpload is read entirely into memory prior to sending data to CC.
func (client *Client) UploadPackage(pkg Package, fileToUpload string) (Package, Warnings, error) {
	body, contentType, err := client.createUploadBuffer(fileToUpload, "bits")
	if err != nil {
		return Package{}, nil, err
	}

	responsePackage := Package{}
	_, warnings, err := client.MakeRequestSendRaw(
		internal.PostPackageBitsRequest,
		internal.Params{"package_guid": pkg.GUID},
		body.Bytes(),
		contentType,
		&responsePackage,
	)

	return responsePackage, warnings, err
}

// CopyPackage copies a package from a source package to a destination package
// Note: source app guid is in URL; dest app guid is in body
func (client *Client) CopyPackage(sourcePkgGUID string, targetAppGUID string) (Package, Warnings, error) {
	var targetPackage Package

	_, warnings, err := client.MakeRequest(RequestParams{
		RequestName: internal.PostPackageRequest,
		Query:       []Query{{Key: SourceGUID, Values: []string{sourcePkgGUID}}},
		RequestBody: map[string]Relationships{
			"relationships": {
				constant.RelationshipTypeApplication: Relationship{GUID: targetAppGUID},
			},
		},
		ResponseBody: &targetPackage,
	})

	return targetPackage, warnings, err
}

func (client *Client) calculateAppBitsRequestSize(matchedResources []Resource, newResourcesLength int64) (int64, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	jsonResources, err := json.Marshal(matchedResources)
	if err != nil {
		return 0, err
	}
	err = form.WriteField("resources", string(jsonResources))
	if err != nil {
		return 0, err
	}
	_, err = form.CreateFormFile("bits", "package.zip")
	if err != nil {
		return 0, err
	}
	err = form.Close()
	if err != nil {
		return 0, err
	}

	return int64(body.Len()) + newResourcesLength, nil
}

func (client *Client) createMultipartBodyAndHeaderForAppBits(matchedResources []Resource, newResources io.Reader, newResourcesLength int64) (string, io.ReadSeeker, <-chan error) {
	writerOutput, writerInput := cloudcontroller.NewPipeBomb()
	form := multipart.NewWriter(writerInput)

	writeErrors := make(chan error)

	go func() {
		defer close(writeErrors)
		defer writerInput.Close()

		jsonResources, err := json.Marshal(matchedResources)
		if err != nil {
			writeErrors <- err
			return
		}

		err = form.WriteField("resources", string(jsonResources))
		if err != nil {
			writeErrors <- err
			return
		}

		writer, err := form.CreateFormFile("bits", "package.zip")
		if err != nil {
			writeErrors <- err
			return
		}

		if newResourcesLength != 0 {
			_, err = io.Copy(writer, newResources)
			if err != nil {
				writeErrors <- err
				return
			}
		}

		err = form.Close()
		if err != nil {
			writeErrors <- err
		}
	}()

	return form.FormDataContentType(), writerOutput, writeErrors
}

func (*Client) createUploadBuffer(path string, paramName string) (bytes.Buffer, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return bytes.Buffer{}, "", err
	}
	defer file.Close()

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return bytes.Buffer{}, "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return bytes.Buffer{}, "", err
	}

	err = writer.Close()

	return body, writer.FormDataContentType(), err
}

func (client *Client) uploadExistingResourcesOnly(packageGUID string, matchedResources []Resource) (Package, Warnings, error) {
	jsonResources, err := json.Marshal(matchedResources)
	if err != nil {
		return Package{}, nil, err
	}

	body := bytes.NewBuffer(nil)
	form := multipart.NewWriter(body)
	err = form.WriteField("resources", string(jsonResources))
	if err != nil {
		return Package{}, nil, err
	}

	err = form.Close()
	if err != nil {
		return Package{}, nil, err
	}

	responsePackage := Package{}

	_, warnings, err := client.MakeRequestSendRaw(
		internal.PostPackageBitsRequest,
		internal.Params{"package_guid": packageGUID},
		body.Bytes(),
		form.FormDataContentType(),
		&responsePackage,
	)

	return responsePackage, warnings, err
}

func (client *Client) uploadNewAndExistingResources(packageGUID string, matchedResources []Resource, newResources io.Reader, newResourcesLength int64) (Package, Warnings, error) {
	contentLength, err := client.calculateAppBitsRequestSize(matchedResources, newResourcesLength)
	if err != nil {
		return Package{}, nil, err
	}

	contentType, body, writeErrors := client.createMultipartBodyAndHeaderForAppBits(matchedResources, newResources, newResourcesLength)

	responseBody := Package{}
	_, warnings, err := client.MakeRequestUploadAsync(
		internal.PostPackageBitsRequest,
		internal.Params{"package_guid": packageGUID},
		contentType,
		body,
		contentLength,
		&responseBody,
		writeErrors,
	)
	return responseBody, warnings, err
}
