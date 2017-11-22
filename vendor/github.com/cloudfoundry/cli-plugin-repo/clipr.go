package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/cloudfoundry/cli-plugin-repo/web"
	"github.com/tedsuo/rata"
	"github.com/unrolled/secure"
	"gopkg.in/yaml.v2"
)

type CLIPR struct {
	Port          int    `short:"p" long:"port" default:"8080" description:"Port that the plugin repo listens on"`
	RepoIndexPath string `short:"f" long:"filepath" default:"repo-index.yml" description:"Path to repo-index file"`
	ForceSSL      bool   `long:"force-ssl"  description:"Force SSL on every request"`
}

func (cmd *CLIPR) Execute(args []string) error {
	logger := os.Stdout //turn this into a logger soon

	var plugins web.PluginsJson

	b, err := ioutil.ReadFile(cmd.RepoIndexPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, &plugins)
	if err != nil {
		return err
	}

	sort.Sort(plugins)

	handlers := map[string]http.Handler{
		Index: http.FileServer(http.Dir("ui")),
		List:  web.NewListPluginsHandler(plugins, logger),
		UI:    http.RedirectHandler("/", http.StatusMovedPermanently),
	}

	router, err := rata.NewRouter(Routes, handlers)

	if err != nil {
		return err
	}

	if cmd.ForceSSL {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect:     true,
			SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		})
		router = secureMiddleware.Handler(router)
	}

	return http.ListenAndServe(cmd.bindAddr(), router)
}

func (cmd *CLIPR) bindAddr() string {
	return fmt.Sprintf(":%d", cmd.Port)
}
