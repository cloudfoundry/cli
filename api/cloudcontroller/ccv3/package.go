package ccv3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

type PackageState string

const (
	PackageStateProcessingUpload PackageState = "PROCESSING_UPLOAD"
	PackageStateReady            PackageState = "READY"
	PackageStateFailed           PackageState = "FAILED"
	PackageStateAwaitingUpload   PackageState = "AWAITING_UPLOAD"
	PackageStateCopying          PackageState = "COPYING"
	PackageStateExpired          PackageState = "EXPIRED"
)

type PackageType string

const (
	PackageTypeBits   PackageType = "bits"
	PackageTypeDocker PackageType = "docker"
)

type Package struct {
	GUID          string               `json:"guid,omitempty"`
	Links         APILinks             `json:"links,omitempty"`
	Relationships PackageRelationships `json:"relationships"`
	State         PackageState         `json:"state,omitempty"`
	Type          PackageType          `json:"type"`
}

type PackageRelationships struct {
	Application Relationship `json:"app"`
}

type UploadLinkNotFoundError struct {
	PackageGUID string
}

func (e UploadLinkNotFoundError) Error() string {
	return fmt.Sprintf("Upload link not found in for package with GUID %s", e.PackageGUID)
}

// GetPackage returns the package with the given GUID.
func (client *Client) GetPackage(guid string) (Package, Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.GetPackageRequest,
		URIParams:   internal.Params{"guid": guid},
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

// CreatePackage creates a package with the given settings, Type and the Space
// must be set.
func (client *Client) CreatePackage(pkg Package) (Package, Warnings, error) {
	bodyBytes, err := json.Marshal(pkg)
	if err != nil {
		return Package{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostPackageRequest,
		Body:        bytes.NewBuffer(bodyBytes),
	})

	var responsePackage Package
	response := cloudcontroller.Response{
		Result: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
}

// UploadPackage uploads a file to a given package's Upload resource.
func (client *Client) UploadPackage(pkg Package, fileToUpload string) (Package, Warnings, error) {
	link, ok := pkg.Links["upload"]
	if !ok {
		return Package{}, nil, UploadLinkNotFoundError{PackageGUID: pkg.GUID}
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
	request.Header.Set("Content-Type", contentType)

	var responsePackage Package
	response := cloudcontroller.Response{
		Result: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
}

func (client *Client) createUploadStream(path string, paramName string) (io.Reader, string, error) {
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

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}
