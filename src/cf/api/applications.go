package api

import (
	"bytes"
	"cf"
	"cf/configuration"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
)

type ApplicationRepository interface {
	FindByName(name string) (app cf.Application, apiErr *ApiError)
	SetEnv(app cf.Application, name string, value string) (apiErr *ApiError)
	Create(newApp cf.Application) (createdApp cf.Application, apiErr *ApiError)
	Delete(app cf.Application) (apiErr *ApiError)
	Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *ApiError)
	Start(app cf.Application) (apiErr *ApiError)
	Stop(app cf.Application) (apiErr *ApiError)
	GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiErr *ApiError)
}

type CloudControllerApplicationRepository struct {
	config    *configuration.Configuration
	apiClient ApiClient
}

func NewCloudControllerApplicationRepository(config *configuration.Configuration, apiClient ApiClient) (repo CloudControllerApplicationRepository) {
	repo.config = config
	repo.apiClient = apiClient
	return
}

func (repo CloudControllerApplicationRepository) FindByName(name string) (app cf.Application, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/spaces/%s/apps?q=name%s&inline-relations-depth=1", repo.config.Target, repo.config.Space.Guid, "%3A"+name)
	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	findResponse := new(ApplicationsApiResponse)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, findResponse)
	if apiErr != nil {
		return
	}

	if len(findResponse.Resources) == 0 {
		apiErr = NewApiErrorWithMessage(fmt.Sprintf("Application %s not found", name))
		return
	}

	res := findResponse.Resources[0]
	path = fmt.Sprintf("%s/v2/apps/%s/summary", repo.config.Target, res.Metadata.Guid)
	request, apiErr = NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	summaryResponse := new(ApplicationSummary)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, summaryResponse)
	if apiErr != nil {
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
		Name:            summaryResponse.Name,
		Guid:            summaryResponse.Guid,
		Instances:       summaryResponse.Instances,
		Memory:          summaryResponse.Memory,
		EnvironmentVars: res.Entity.EnvironmentJson,
		Urls:            urls,
	}

	return
}

func (repo CloudControllerApplicationRepository) SetEnv(app cf.Application, name string, value string) (apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	data := fmt.Sprintf(`{"environment_json":{"%s":"%s"}}`, name, value)
	request, apiErr := NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Create(newApp cf.Application) (createdApp cf.Application, apiErr *ApiError) {
	apiErr = validateApplication(newApp)
	if apiErr != nil {
		return
	}

	buildpackUrl := stringOrNull(newApp.BuildpackUrl)
	stackGuid := stringOrNull(newApp.Stack.Guid)

	path := fmt.Sprintf("%s/v2/apps", repo.config.Target)
	data := fmt.Sprintf(
		`{"space_guid":"%s","name":"%s","instances":%d,"buildpack":%s,"command":null,"memory":%d,"stack_guid":%s}`,
		repo.config.Space.Guid, newApp.Name, newApp.Instances, buildpackUrl, newApp.Memory, stackGuid,
	)
	request, apiErr := NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(data))
	if apiErr != nil {
		return
	}

	resource := new(Resource)
	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, resource)
	if apiErr != nil {
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

func (repo CloudControllerApplicationRepository) Delete(app cf.Application) (apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s?recursive=true", repo.config.Target, app.Guid)
	request, apiErr := NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Upload(app cf.Application, zipBuffer *bytes.Buffer) (apiErr *ApiError) {
	url := fmt.Sprintf("%s/v2/apps/%s/bits", repo.config.Target, app.Guid)

	body, boundary, err := createApplicationUploadBody(zipBuffer)
	if err != nil {
		apiErr = NewApiErrorWithError("Error creating upload", err)
		return
	}

	request, apiErr := NewRequest("PUT", url, repo.config.AccessToken, body)
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	request.Header.Set("Content-Type", contentType)
	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)
	return
}

func (repo CloudControllerApplicationRepository) Start(app cf.Application) (apiErr *ApiError) {
	return repo.changeApplicationState(app, "STARTED")
}

func (repo CloudControllerApplicationRepository) Stop(app cf.Application) (apiErr *ApiError) {
	return repo.changeApplicationState(app, "STOPPED")
}

type InstancesApiResponse map[string]InstanceApiResponse

type InstanceApiResponse struct {
	State string
}

func (repo CloudControllerApplicationRepository) GetInstances(app cf.Application) (instances []cf.ApplicationInstance, apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s/instances", repo.config.Target, app.Guid)
	request, apiErr := NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiErr != nil {
		return
	}

	apiResponse := InstancesApiResponse{}

	apiErr = repo.apiClient.PerformRequestAndParseResponse(request, &apiResponse)
	if apiErr != nil {
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

func (repo CloudControllerApplicationRepository) changeApplicationState(app cf.Application, state string) (apiErr *ApiError) {
	path := fmt.Sprintf("%s/v2/apps/%s", repo.config.Target, app.Guid)
	body := fmt.Sprintf(`{"console":true,"state":"%s"}`, state)
	request, apiErr := NewRequest("PUT", path, repo.config.AccessToken, strings.NewReader(body))

	if apiErr != nil {
		return
	}

	apiErr = repo.apiClient.PerformRequest(request)
	return
}

func validateApplication(app cf.Application) (apiErr *ApiError) {
	reg := regexp.MustCompile("^[0-9a-zA-Z\\-_]*$")
	if !reg.MatchString(app.Name) {
		apiErr = NewApiErrorWithMessage("Application name is invalid. Name can only contain letters, numbers, underscores and hyphens.")
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
