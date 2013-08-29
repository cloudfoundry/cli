package api

import (
	"archive/zip"
	"bytes"
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ApplicationRepository interface {
	FindAll(config *configuration.Configuration) (apps []cf.Application, err error)
	FindByName(config *configuration.Configuration, name string) (app cf.Application, err error)
	SetEnv(config *configuration.Configuration, app cf.Application, name string, value string) (err error)
	Create(config *configuration.Configuration, newApp cf.Application) (createdApp cf.Application, err error)
	Delete(config *configuration.Configuration, app cf.Application) (err error)
	Upload(config *configuration.Configuration, app cf.Application) (err error)
	Start(config *configuration.Configuration, app cf.Application) (err error)
	Stop(config *configuration.Configuration, app cf.Application) (err error)
	GetInstances(config *configuration.Configuration, app cf.Application) (instances []cf.ApplicationInstance, err error)
}

type CloudControllerApplicationRepository struct {
}

func (repo CloudControllerApplicationRepository) FindAll(config *configuration.Configuration) (apps []cf.Application, err error) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?inline-relations-depth=2", config.Target, config.Space.Guid)
	request, err := NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	response := new(ApplicationsApiResponse)
	err = PerformRequestAndParseResponse(request, response)
	if err != nil {
		return
	}

	for _, res := range response.Resources {
		urls := []string{}
		for _, routeRes := range res.Entity.Routes {
			routeEntity := routeRes.Entity
			domainEntity := routeEntity.Domain.Entity
			urls = append(urls, fmt.Sprintf("%s.%s", routeEntity.Host, domainEntity.Name))
		}

		apps = append(apps, cf.Application{
			Name:      res.Entity.Name,
			Guid:      res.Metadata.Guid,
			State:     strings.ToLower(res.Entity.State),
			Instances: res.Entity.Instances,
			Memory:    res.Entity.Memory,
			Urls:      urls,
		})
	}

	return
}

func (repo CloudControllerApplicationRepository) FindByName(config *configuration.Configuration, name string) (app cf.Application, err error) {
	apps, err := repo.FindAll(config)
	lowerName := strings.ToLower(name)
	if err != nil {
		return
	}

	for _, a := range apps {
		if strings.ToLower(a.Name) == lowerName {
			return a, nil
		}
	}

	err = errors.New("Application not found")
	return
}

func (repo CloudControllerApplicationRepository) SetEnv(config *configuration.Configuration, app cf.Application, name string, value string) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s", config.Target, app.Guid)
	data := fmt.Sprintf(`{"environment_json":{"%s":"%s"}}`, name, value)
	request, err := NewAuthorizedRequest("PUT", path, config.AccessToken, strings.NewReader(data))
	if err != nil {
		return
	}

	err = PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Create(config *configuration.Configuration, newApp cf.Application) (createdApp cf.Application, err error) {
	err = validateApplication(newApp)
	if err != nil {
		return
	}

	path := fmt.Sprintf("%s/v2/apps", config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":null,"command":null,"memory":%d,"stack_guid":null}`,
		config.Space.Guid, newApp.Name, newApp.Instances, newApp.Memory,
	)
	request, err := NewAuthorizedRequest("POST", path, config.AccessToken, strings.NewReader(data))
	if err != nil {
		return
	}

	resource := new(Resource)
	err = PerformRequestAndParseResponse(request, resource)

	if err != nil {
		return
	}

	createdApp.Guid = resource.Metadata.Guid
	createdApp.Name = resource.Entity.Name
	return
}

func (repo CloudControllerApplicationRepository) Delete(config *configuration.Configuration, app cf.Application) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", config.Target, app.Guid)
	request, err := NewAuthorizedRequest("DELETE", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	err = PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Upload(config *configuration.Configuration, app cf.Application) (err error) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", config.Target, app.Guid)
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	zipBuffer, err := ZipApplication(dir)
	if err != nil {
		return
	}

	body, boundary, err := createApplicationUploadBody(zipBuffer)
	if err != nil {
		return
	}

	request, err := NewAuthorizedRequest("PUT", url, config.AccessToken, body)
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	request.Header.Set("Content-Type", contentType)
	if err != nil {
		return
	}

	err = PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Start(config *configuration.Configuration, app cf.Application) (err error) {
	return changeApplicationState(config, app, "STARTED")
}

func (repo CloudControllerApplicationRepository) Stop(config *configuration.Configuration, app cf.Application) (err error) {
	return changeApplicationState(config, app, "STOPPED")
}

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
}

func (repo CloudControllerApplicationRepository) GetInstances(config *configuration.Configuration, app cf.Application) (instances []cf.ApplicationInstance, err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", config.Target, app.Guid)
	request, err := NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	apiResponse := InstancesApiResponse{}

	err = PerformRequestAndParseResponse(request, &apiResponse)
	if err != nil {
		return
	}

	instances = make([]cf.ApplicationInstance, len(apiResponse), len(apiResponse))
	for k, v := range apiResponse {
		index, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		instances[index] = cf.ApplicationInstance{State: cf.InstanceState(strings.ToLower(v.State))}
	}
	return
}

func changeApplicationState(config *configuration.Configuration, app cf.Application, state string) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s", config.Target, app.Guid)
	body := fmt.Sprintf(`{"console":true,"state":"%s"}`, state)
	request, err := NewAuthorizedRequest("PUT", path, config.AccessToken, strings.NewReader(body))

	if err != nil {
		return
	}

	return PerformRequest(request)
}

func validateApplication(app cf.Application) (err error) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		err = errors.New("Application name is invalid. Name can only contain letters, numbers, underscores and hyphens.")
	}

	return
}

func ZipApplication(dirToZip string) (zipBuffer *bytes.Buffer, err error) {
	zipBuffer = new(bytes.Buffer)
	writer := zip.NewWriter(zipBuffer)

	addFileToZip := func(path string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		if f.IsDir() {
			return
		}

		fileName := strings.TrimPrefix(path, dirToZip+"/")
		zipFile, err := writer.Create(fileName)
		if err != nil {
			return
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return
		}

		_, err = zipFile.Write(content)
		if err != nil {
			return
		}

		return
	}

	err = filepath.Walk(dirToZip, addFileToZip)

	if err != nil {
		return
	}

	err = writer.Close()
	return
}

func createApplicationUploadBody(zipBuffer *bytes.Buffer) (body *bytes.Buffer, boundary string, err error) {
	body = new(bytes.Buffer)

	writer := multipart.NewWriter(body)
	boundary = writer.Boundary()

	part, err := writer.CreateFormField("resources")
	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewBufferString("[]"))
	if err != nil {
		return
	}

	part, err = createZipPartWriter(zipBuffer, writer)
	if err != nil {
		return
	}

	_, err = io.Copy(part, zipBuffer)
	if err != nil {
		return
	}

	err = writer.Close()
	return
}

func createZipPartWriter(zipBuffer *bytes.Buffer, writer *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", zipBuffer.Len()))
	h.Set("Content-Transfer-Encoding", "binary")
	return writer.CreatePart(h)
}
