package helpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type Plugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type PluginRepository struct {
	Plugins []Plugin `json:"plugins"`
}

func NewPluginRepositoryServer(pluginRepo PluginRepository) (*Server, string) {
	return configurePluginRepositoryServer(NewServer(), pluginRepo)
}

func NewPluginRepositoryTLSServer(pluginRepo PluginRepository) (*Server, string) {
	return configurePluginRepositoryServer(NewTLSServer(), pluginRepo)
}

func configurePluginRepositoryServer(server *Server, pluginRepo PluginRepository) (*Server, string) {
	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)

	jsonBytes, err := json.Marshal(pluginRepo)
	Expect(err).ToNot(HaveOccurred())

	server.AppendHandlers(
		RespondWith(http.StatusOK, string(jsonBytes)),
		RespondWith(http.StatusOK, string(jsonBytes)),
		RespondWith(http.StatusOK, string(jsonBytes)),
	)

	return server, server.URL()
}
