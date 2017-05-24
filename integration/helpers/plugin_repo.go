package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/util"
	"code.cloudfoundry.org/cli/util/generic"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

type Binary struct {
	Checksum string `json:"checksum"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type Plugin struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Binaries []Binary `json:"binaries"`
}

type PluginRepository struct {
	Plugins []Plugin `json:"plugins"`
}

type PluginRepositoryServerWithPlugin struct {
	server     *Server
	pluginPath string
}

func NewPluginRepositoryServer(pluginRepo PluginRepository) *Server {
	return configurePluginRepositoryServer(NewTLSServer(), pluginRepo)
}

func NewPluginRepositoryServerWithPlugin(pluginName string, version string, platform string, shouldCalculateChecksum bool) *PluginRepositoryServerWithPlugin {
	pluginRepoServer := PluginRepositoryServerWithPlugin{}

	pluginRepoServer.Init(pluginName, version, platform, shouldCalculateChecksum)

	return &pluginRepoServer
}

func (pluginRepoServer *PluginRepositoryServerWithPlugin) Init(pluginName string, version string, platform string, shouldCalculateChecksum bool) {
	pluginPath := BuildConfigurablePlugin("configurable_plugin", pluginName, version,
		[]PluginCommand{
			{Name: "some-command", Help: "some-command-help"},
		},
	)

	repoServer := NewServer()

	pluginRepoServer.server = repoServer
	pluginRepoServer.pluginPath = pluginPath

	var (
		checksum []byte
		err      error
	)

	if shouldCalculateChecksum {
		checksum, err = util.NewSha1Checksum(pluginPath).ComputeFileSha1()
		Expect(err).NotTo(HaveOccurred())
	}

	baseFile := fmt.Sprintf("/%s", generic.ExecutableFilename(filepath.Base(pluginPath)))
	downloadURL := fmt.Sprintf("%s%s", repoServer.URL(), baseFile)
	pluginRepo := PluginRepository{
		Plugins: []Plugin{
			{
				Name:    pluginName,
				Version: version,
				Binaries: []Binary{
					{
						Checksum: fmt.Sprintf("%x", checksum),
						Platform: platform,
						URL:      downloadURL,
					},
				},
			},
		}}

	// Suppresses ginkgo server logs
	repoServer.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)

	jsonBytes, err := json.Marshal(pluginRepo)
	Expect(err).ToNot(HaveOccurred())

	pluginData, err := ioutil.ReadFile(pluginPath)
	Expect(err).ToNot(HaveOccurred())

	repoServer.AppendHandlers(
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/list"),
			RespondWith(http.StatusOK, jsonBytes),
		),
		CombineHandlers(
			VerifyRequest(http.MethodGet, "/list"),
			RespondWith(http.StatusOK, jsonBytes),
		),
		CombineHandlers(
			VerifyRequest(http.MethodGet, baseFile),
			RespondWith(http.StatusOK, pluginData),
		),
	)
}

func (pluginRepoServer *PluginRepositoryServerWithPlugin) PluginSize() int64 {
	fileinfo, err := os.Stat(pluginRepoServer.pluginPath)
	Expect(err).NotTo(HaveOccurred())
	return fileinfo.Size()
}

func (pluginRepoServer *PluginRepositoryServerWithPlugin) URL() string {
	return pluginRepoServer.server.URL()
}

func (pluginRepoServer *PluginRepositoryServerWithPlugin) Cleanup() {
	pluginRepoServer.server.Close()
	Expect(os.RemoveAll(filepath.Dir(pluginRepoServer.pluginPath))).NotTo(HaveOccurred())
}

func NewPluginRepositoryTLSServer(pluginRepo PluginRepository) *Server {
	return configurePluginRepositoryServer(NewTLSServer(), pluginRepo)
}

func configurePluginRepositoryServer(server *Server, pluginRepo PluginRepository) *Server {
	// Suppresses ginkgo server logs
	server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)

	jsonBytes, err := json.Marshal(pluginRepo)
	Expect(err).ToNot(HaveOccurred())

	server.AppendHandlers(
		RespondWith(http.StatusOK, string(jsonBytes)),
		RespondWith(http.StatusOK, string(jsonBytes)),
		RespondWith(http.StatusOK, string(jsonBytes)),
		RespondWith(http.StatusOK, string(jsonBytes)),
	)

	return server
}
