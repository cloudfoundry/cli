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
		DecodeJSONResponseInto: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
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
		DecodeJSONResponseInto: &responsePackage,
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

// UploadApplicationPackage uploads the newResources and a list of existing
// resources to the cloud controller. An updated package is returned. The
// function will act differently given the following Readers:
//   - io.ReadSeeker: Will function properly on retry.
//   - io.Reader: Will return a ccerror.PipeSeekError on retry.
//   - nil: Will not add the "application" section to the request. The newResourcesLength is ignored in this case.
//
// Note: In order to determine if package creation is successful, poll the
// Package's state field for more information.
func (client *Client) UploadBitsPackage(pkg Package, matchedResources []Resource, newResources io.Reader, newResourcesLength int64) (Package, Warnings, error) {
	link, ok := pkg.Links["upload"]
	if !ok {
		return Package{}, nil, ccerror.UploadLinkNotFoundError{PackageGUID: pkg.GUID}
	}

	if matchedResources == nil {
		return Package{}, nil, ccerror.NilObjectError{Object: "matchedResources"}
	}

	if newResources == nil {
		return client.uploadExistingResourcesOnly(link, matchedResources)
	}

	return client.uploadNewAndExistingResources(link, matchedResources, newResources, newResourcesLength)
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
		DecodeJSONResponseInto: &responsePackage,
	}
	err = client.connection.Make(request, &response)

	return responsePackage, response.Warnings, err
}

func (*Client) calculateAppBitsRequestSize(matchedResources []Resource, newResourcesLength int64) (int64, error) {
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

func (*Client) createMultipartBodyAndHeaderForAppBits(matchedResources []Resource, newResources io.Reader, newResourcesLength int64) (string, io.ReadSeeker, <-chan error) {
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

func (client *Client) uploadAsynchronously(request *cloudcontroller.Request, writeErrors <-chan error) (Package, Warnings, error) {
	var pkg Package
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &pkg,
	}

	httpErrors := make(chan error)

	go func() {
		defer close(httpErrors)

		err := client.connection.Make(request, &response)
		if err != nil {
			httpErrors <- err
		}
	}()

	// The following section makes the following assumptions:
	// 1) If an error occurs during file reading, an EOF is sent to the request
	// object. Thus ending the request transfer.
	// 2) If an error occurs during request transfer, an EOF is sent to the pipe.
	// Thus ending the writing routine.
	var firstError error
	var writeClosed, httpClosed bool

	for {
		select {
		case writeErr, ok := <-writeErrors:
			if !ok {
				writeClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = writeErr
			}
		case httpErr, ok := <-httpErrors:
			if !ok {
				httpClosed = true
				break // for select
			}
			if firstError == nil {
				firstError = httpErr
			}
		}

		if writeClosed && httpClosed {
			break // for for
		}
	}

	return pkg, response.Warnings, firstError
}

func (client *Client) uploadExistingResourcesOnly(uploadLink APILink, matchedResources []Resource) (Package, Warnings, error) {
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

	request, err := client.newHTTPRequest(requestOptions{
		URL:    uploadLink.HREF,
		Method: uploadLink.Method,
		Body:   bytes.NewReader(body.Bytes()),
	})
	if err != nil {
		return Package{}, nil, err
	}

	request.Header.Set("Content-Type", form.FormDataContentType())

	var pkg Package
	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &pkg,
	}

	err = client.connection.Make(request, &response)
	return pkg, response.Warnings, err
}

func (client *Client) uploadNewAndExistingResources(uploadLink APILink, matchedResources []Resource, newResources io.Reader, newResourcesLength int64) (Package, Warnings, error) {
	contentLength, err := client.calculateAppBitsRequestSize(matchedResources, newResourcesLength)
	if err != nil {
		return Package{}, nil, err
	}

	contentType, body, writeErrors := client.createMultipartBodyAndHeaderForAppBits(matchedResources, newResources, newResourcesLength)

	// This request uses URL/Method instead of an internal RequestName to support
	// the possibility of external bit services.
	request, err := client.newHTTPRequest(requestOptions{
		URL:    uploadLink.HREF,
		Method: uploadLink.Method,
		Body:   body,
	})
	if err != nil {
		return Package{}, nil, err
	}

	request.Header.Set("Content-Type", contentType)
	request.ContentLength = contentLength

	return client.uploadAsynchronously(request, writeErrors)
}
