package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
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
	Upload(config *configuration.Configuration, app cf.Application, zipBuffer *bytes.Buffer) (err error)
	Start(config *configuration.Configuration, app cf.Application) (err error)
	Stop(config *configuration.Configuration, app cf.Application) (err error)
	GetInstances(config *configuration.Configuration, app cf.Application) (instances []cf.ApplicationInstance, errorCode int, err error)
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
	_, err = PerformRequestAndParseResponse(request, response)
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
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", config.Target, config.Space.Guid, "%3A"+name)
	request, err := NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	findResponse := new(ApplicationsApiResponse)
	_, err = PerformRequestAndParseResponse(request, findResponse)
	if err != nil {
		return
	}

	if len(findResponse.Resources) == 0 {
		err = errors.New(fmt.Sprintf("Application %s not found", name))
		return
	}

	res := findResponse.Resources[0]
	path = fmt.Sprintf("%s/v2/apps/%s/summary", config.Target, res.Metadata.Guid)
	request, err = NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	summaryResponse := new(ApplicationSummary)
	_, err = PerformRequestAndParseResponse(request, summaryResponse)
	if err != nil {
		return
	}

	urls := []string{}
	// This is a little wonky but we made a concious effort
	// to keep the domain very separate from the API repsonses
	// to maintain flexibility.
	domainRoute := cf.Route{}
	for _, route := range summaryResponse.Routes {
		domainRoute.Domain = cf.Domain{Name: route.Domain.Name}
		domainRoute.Host = route.Host
		urls = append(urls, domainRoute.URL())
	}

	app = cf.Application{
		Name:      summaryResponse.Name,
		Guid:      summaryResponse.Guid,
		Instances: summaryResponse.Instances,
		Memory:    summaryResponse.Memory,
		Urls:      urls,
	}

	return
}

func (repo CloudControllerApplicationRepository) SetEnv(config *configuration.Configuration, app cf.Application, name string, value string) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s", config.Target, app.Guid)
	data := fmt.Sprintf(`{"environment_json":{"%s":"%s"}}`, name, value)
	request, err := NewAuthorizedRequest("PUT", path, config.AccessToken, strings.NewReader(data))
	if err != nil {
		return
	}

	_, err = PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Create(config *configuration.Configuration, newApp cf.Application) (createdApp cf.Application, err error) {
	err = validateApplication(newApp)
	if err != nil {
		return
	}

	buildpackUrl := stringOrNull(newApp.BuildpackUrl)
	stackGuid := stringOrNull(newApp.Stack.Guid)

	path := fmt.Sprintf("%s/v2/apps", config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":%s,"command":null,"memory":%d,"stack_guid":%s}`,
		config.Space.Guid, newApp.Name, newApp.Instances, buildpackUrl, newApp.Memory, stackGuid,
	)
	request, err := NewAuthorizedRequest("POST", path, config.AccessToken, strings.NewReader(data))
	if err != nil {
		return
	}

	resource := new(Resource)
	_, err = PerformRequestAndParseResponse(request, resource)

	if err != nil {
		return
	}

	createdApp.Guid = resource.Metadata.Guid
	createdApp.Name = resource.Entity.Name
	return
}

func stringOrNull(s string) string {
	if s == "" {
		return "null"
	}

	return fmt.Sprintf(`"%s"`, s)
}

func (repo CloudControllerApplicationRepository) Delete(config *configuration.Configuration, app cf.Application) (err error) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", config.Target, app.Guid)
	request, err := NewAuthorizedRequest("DELETE", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	_, err = PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Upload(config *configuration.Configuration, app cf.Application, zipBuffer *bytes.Buffer) (err error) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", config.Target, app.Guid)

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

	_, err = PerformRequest(request)
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

func (repo CloudControllerApplicationRepository) GetInstances(config *configuration.Configuration, app cf.Application) (instances []cf.ApplicationInstance, errorCode int, err error) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", config.Target, app.Guid)
	request, err := NewAuthorizedRequest("GET", path, config.AccessToken, nil)
	if err != nil {
		return
	}

	apiResponse := InstancesApiResponse{}

	errorCode, err = PerformRequestAndParseResponse(request, &apiResponse)
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

	_, err = PerformRequest(request)
	return
}

func validateApplication(app cf.Application) (err error) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		err = errors.New("Application name is invalid. Name can only contain letters, numbers, underscores and hyphens.")
	}

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
