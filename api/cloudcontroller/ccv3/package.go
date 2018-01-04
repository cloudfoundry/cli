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

type Package struct {
	GUID           string
	CreatedAt      string
	Links          APILinks
	Relationships  Relationships
	State          constant.PackageState
	Type           constant.PackageType
	DockerImage    string
	DockerUsername string
	DockerPassword string
}

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
	if err := json.Unmarshal(data, &ccPackage); err != nil {
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

// GetPackage returns the package with the given GUID.
func (client *Client) GetPackage(packageGUID string) (Package, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPackageRequest,
		URIParams:   internal.Params{"package_guid": packageGUID},
	})
	if err != nil {
		return Package{}, nil, err
	}

	var responsePackage Package
	response := cloudcontroller.Response{
		Result: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
}

// CreatePackage creates a package with the given settings, Type and the
// ApplicationRelationship must be set.
func (client *Client) CreatePackage(pkg Package) (Package, Warnings, error) {
	bodyBytes, err := json.Marshal(pkg)
	if err != nil {
		return Package{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostPackageRequest,
		Body:        bytes.NewReader(bodyBytes),
	})
	if err != nil {
		return Package{}, nil, err
	}

	var responsePackage Package
	response := cloudcontroller.Response{
		Result: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
}

// UploadPackage uploads a file to a given package's Upload resource. Note:
// fileToUpload is read entirely into memory prior to sending data to CC.
func (client *Client) UploadPackage(pkg Package, fileToUpload string) (Package, Warnings, error) {
	link, ok := pkg.Links["upload"]
	if !ok {
		return Package{}, nil, ccerror.UploadLinkNotFoundError{PackageGUID: pkg.GUID}
	}

	body, contentType, err := client.createUploadStream(fileToUpload, "bits")
	if err != nil {
		return Package{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		URL:    link.HREF,
		Method: link.Method,
		Body:   body,
	})
	if err != nil {
		return Package{}, nil, err
	}

	request.Header.Set("Content-Type", contentType)

	var responsePackage Package
	response := cloudcontroller.Response{
		Result: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
}

// GetPackages returns the list of packages.
func (client *Client) GetPackages(query ...Query) ([]Package, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPackagesRequest,
		Query:       query,
	})
	if err != nil {
		return nil, nil, err
	}

	var fullPackagesList []Package
	warnings, err := client.paginate(request, Package{}, func(item interface{}) error {
		if pkg, ok := item.(Package); ok {
			fullPackagesList = append(fullPackagesList, pkg)
		} else {
			return ccerror.UnknownObjectInListError{
				Expected:   Package{},
				Unexpected: item,
			}
		}
		return nil
	})

	return fullPackagesList, warnings, err
}

func (*Client) createUploadStream(path string, paramName string) (io.ReadSeeker, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", err
	}

	err = writer.Close()

	return bytes.NewReader(body.Bytes()), writer.FormDataContentType(), err
}
