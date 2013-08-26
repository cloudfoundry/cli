package api

import (
	"archive/zip"
	"bytes"
	"cf"
	"cf/configuration"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var multipleAppsEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "GET"
	pathMatches := request.URL.Path == "/v2/spaces/my-space-guid/apps"
	queryStringMatches := strings.Contains(request.RequestURI, "?inline-relations-depth=2")
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	if !(acceptHeaderMatches && methodMatches && pathMatches && queryStringMatches && authMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonResponse := `
{
  "resources": [
    {
      "metadata": {
        "guid": "app1-guid"
      },
      "entity": {
        "name": "App1",
        "memory": 256,
        "instances": 1,
        "state": "STOPPED",
        "routes": [
      	  {
      	    "metadata": {
      	      "guid": "app1-route-guid"
      	    },
      	    "entity": {
      	      "host": "app1",
      	      "domain": {
      	      	"metadata": {
      	      	  "guid": "domain1-guid"
      	      	},
      	      	"entity": {
      	      	  "name": "cfapps.io"
      	      	}
      	      }
      	    }
      	  }
        ]
      }
    },
    {
      "metadata": {
        "guid": "app2-guid"
      },
      "entity": {
        "name": "App2",
        "memory": 512,
        "instances": 2,
        "state": "STARTED",
        "routes": [
      	  {
      	    "metadata": {
      	      "guid": "app2-route-guid"
      	    },
      	    "entity": {
      	      "host": "app2",
      	      "domain": {
      	      	"metadata": {
      	      	  "guid": "domain1-guid"
      	      	},
      	      	"entity": {
      	      	  "name": "cfapps.io"
      	      	}
      	      }
      	    }
      	  }
        ]
      }
    }
  ]
}`
	fmt.Fprintln(writer, jsonResponse)
}

func TestApplicationsFindAll(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleAppsEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	apps, err := repo.FindAll(config)
	assert.NoError(t, err)
	assert.Equal(t, len(apps), 2)

	app := apps[0]
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")
	assert.Equal(t, app.State, "stopped")
	assert.Equal(t, app.Instances, 1)
	assert.Equal(t, app.Memory, 256)
	assert.Equal(t, len(app.Urls), 1)
	assert.Equal(t, app.Urls[0], "app1.cfapps.io")

	app = apps[1]
	assert.Equal(t, app.Guid, "app2-guid")

}

func TestFindByName(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(multipleAppsEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Name: "my-space", Guid: "my-space-guid"},
	}

	app, err := repo.FindByName(config, "App1")
	assert.NoError(t, err)
	assert.Equal(t, app.Name, "App1")
	assert.Equal(t, app.Guid, "app1-guid")

	app, err = repo.FindByName(config, "app1")
	assert.NoError(t, err)
	assert.Equal(t, app.Guid, "app1-guid")

	app, err = repo.FindByName(config, "app that does not exist")
	assert.Error(t, err)
}

var setEnvEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "PUT"
	pathMatches := request.URL.Path == "/v2/apps/app1-guid"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
	bodyMatches := string(bodyBytes) == `{"environment_json":{"DATABASE_URL":"mysql://example.com/my-db"}}`

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches && bodyMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
	}
}

func TestSetEnv(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(setEnvEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Guid: "app1-guid", Name: "App1"}

	err := repo.SetEnv(config, app, "DATABASE_URL", "mysql://example.com/my-db")

	assert.NoError(t, err)
}

var createApplicationEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "POST"
	pathMatches := request.URL.Path == "/v2/apps"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
	expectedBody := `{"space_guid":"my-space-guid","name":"my-cool-app","instances":1,"buildpack":null,"command":null,"memory":256,"stack_guid":null}`
	bodyMatches := string(bodyBytes) == expectedBody

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches && bodyMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
		jsonResponse := `
{
    "metadata": {
        "guid": "my-cool-app-guid"
    },
    "entity": {
        "name": "my-cool-app"
    }
}`
		fmt.Fprintln(writer, jsonResponse)

	}
}

var alwaysSuccessfulEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintln(writer, "{}")
}

func TestCreateApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(createApplicationEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
		Space:       cf.Space{Guid: "my-space-guid"},
	}

	newApp := cf.Application{Name: "my-cool-app"}

	createdApp, err := repo.Create(config, newApp)
	assert.NoError(t, err)

	assert.Equal(t, createdApp, cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"})
}

func TestCreateRejectsInproperNames(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(alwaysSuccessfulEndpoint))
	defer ts.Close()

	config := &configuration.Configuration{Target: ts.URL}
	repo := CloudControllerApplicationRepository{}

	createdApp, err := repo.Create(config, cf.Application{Name: "name with space"})
	assert.Equal(t, createdApp, cf.Application{})
	assert.Contains(t, err.Error(), "Application name is invalid")

	_, err = repo.Create(config, cf.Application{Name: "name-with-inv@lid-chars!"})
	assert.Error(t, err)

	_, err = repo.Create(config, cf.Application{Name: "Valid-Name"})
	assert.NoError(t, err)

	_, err = repo.Create(config, cf.Application{Name: "name_with_numbers_2"})
	assert.NoError(t, err)
}

var deleteApplicationEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "DELETE"
	pathMatches := request.URL.Path == "/v2/apps/my-cool-app-guid"
	queryParamsMatch := strings.Contains(request.RequestURI, "?recursive=true")
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	if !(acceptHeaderMatches && methodMatches && pathMatches && queryParamsMatch && authMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
		fmt.Fprintln(writer, "")
	}
}

func TestDeleteApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(deleteApplicationEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Delete(config, app)
	assert.NoError(t, err)
}

var uploadApplicationEndpoint = func(writer http.ResponseWriter, request *http.Request) {
	bodyBytes, err := ioutil.ReadAll(request.Body)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	acceptHeaderMatches := request.Header.Get("accept") == "application/json"
	methodMatches := request.Method == "PUT"
	pathMatches := request.URL.Path == "/v2/apps/my-cool-app-guid/bits"
	authMatches := request.Header.Get("authorization") == "BEARER my_access_token"

	bodyString := string(bodyBytes)
	zipAttachmentContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="application"; filename="application.zip"`)
	zipAttachmentContentTypeMatches := strings.Contains(bodyString, `Content-Type: application/zip`)
	zipAttachmentContentTransferEncodingMatches := strings.Contains(bodyString, `Content-Transfer-Encoding: binary`)
	zipAttachmentContentLengthPresent := strings.Contains(bodyString, `Content-Length:`)

	resourcesContentDispositionMatches := strings.Contains(bodyString, `Content-Disposition: form-data; name="resources"`)

	bodyMatches := zipAttachmentContentDispositionMatches &&
		zipAttachmentContentTypeMatches &&
		zipAttachmentContentTransferEncodingMatches &&
		zipAttachmentContentLengthPresent &&
		resourcesContentDispositionMatches

	if !(acceptHeaderMatches && methodMatches && pathMatches && authMatches && bodyMatches) {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusCreated)
	}
}

func TestUploadApplication(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(uploadApplicationEndpoint))
	defer ts.Close()

	repo := CloudControllerApplicationRepository{}
	config := &configuration.Configuration{
		AccessToken: "BEARER my_access_token",
		Target:      ts.URL,
	}

	app := cf.Application{Name: "my-cool-app", Guid: "my-cool-app-guid"}

	err := repo.Upload(config, app)
	assert.NoError(t, err)
}

func TestZipApplication(t *testing.T) {
	dir, err := os.Getwd()
	assert.NoError(t, err)

	zipFile, err := zipApplication(filepath.Clean(dir + "/../../fixtures/zip/"))
	assert.NoError(t, err)

	byteReader := bytes.NewReader(zipFile.Bytes())
	reader, err := zip.NewReader(byteReader, int64(byteReader.Len()))
	assert.NoError(t, err)

	readFile := func(index int) (string, string) {
		buf := &bytes.Buffer{}
		file := reader.File[index]
		fReader, err := file.Open()
		_, err = io.Copy(buf, fReader)

		assert.NoError(t, err)

		return file.Name, string(buf.Bytes())
	}

	name, contents := readFile(0)
	assert.Equal(t, name, "foo.txt")
	assert.Equal(t, contents, "This is a simple text file.")

	name, contents = readFile(1)
	assert.Equal(t, name, "subDir/bar.txt")
	assert.Equal(t, contents, "I am in a subdirectory.")
}
