package api_test

import (
	"bytes"
	"cf"
	. "cf/api"
	"cf/configuration"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	messagetesthelpers "github.com/cloudfoundry/loggregatorlib/logmessage/testhelpers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testhelpers"
	"testing"
)

var logsEndpoint = func(message *logmessage.Message) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authMatches := request.Header.Get("authorization") == "BEARER my_access_token"
		methodMatches := request.Method == "GET"

		path := "/dump/?app=my-app-guid"

		paths := strings.Split(path, "?")
		pathMatches := request.URL.Path == paths[0]
		if len(paths) > 1 {
			queryStringMatches := strings.Contains(request.RequestURI, paths[1])
			pathMatches = pathMatches && queryStringMatches
		}

		if !(authMatches && methodMatches && pathMatches) {
			fmt.Printf("One of the matchers did not match. Auth [%t] Method [%t] Path [%t]", authMatches, methodMatches, pathMatches)

			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var response *bytes.Buffer
		response = new(bytes.Buffer)
		logmessage.DumpMessage(*message, response)

		writer.Write(response.Bytes())
	}
}

func TestRecentLogsFor(t *testing.T) {
	expectedMessage := messagetesthelpers.MarshalledLogMessage(t, "My message", "my-app-id")
	message, err := logmessage.ParseMessage(expectedMessage)
	assert.NoError(t, err)

	ts := httptest.NewTLSServer(http.HandlerFunc(logsEndpoint(message)))
	defer ts.Close()

	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	config := &configuration.Configuration{AccessToken: "BEARER my_access_token", Target: ts.URL}
	client := NewApiClient(&testhelpers.FakeAuthenticator{})
	redirectHandler := func(hostname string) string { return hostname }
	logsRepo := NewLoggregatorLogsRepository(config, client, redirectHandler)

	logs, err := logsRepo.RecentLogsFor(app)
	assert.NoError(t, err)

	assert.Equal(t, len(logs), 1)
	actualMessage, err := proto.Marshal(logs[0])
	assert.NoError(t, err)
	assert.Equal(t, actualMessage, expectedMessage)
}

func TestLoggregatorHost(t *testing.T) {
	apiHost := "https://api.run.pivotal.io"
	loggregatorHost := LoggregatorHost(apiHost)

	assert.Equal(t, loggregatorHost, "https://loggregator.run.pivotal.io")
}
