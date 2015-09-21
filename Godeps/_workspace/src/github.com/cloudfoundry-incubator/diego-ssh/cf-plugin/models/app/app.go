package app

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models"
	"github.com/cloudfoundry/cli/plugin"
)

//go:generate counterfeiter -o app_fakes/fake_app_factory.go . AppFactory
type AppFactory interface {
	Get(string) (App, error)
	SetBool(anApp App, key string, value bool) error
}

type appFactory struct {
	cli  plugin.CliConnection
	curl models.Curler
}

func NewAppFactory(cli plugin.CliConnection, curl models.Curler) AppFactory {
	return &appFactory{cli: cli, curl: curl}
}

type App struct {
	Guid      string
	EnableSSH bool
	Diego     bool
	State     string
}

type metadata struct {
	Guid string `json:"guid"`
}

type entity struct {
	EnableSSH bool   `json:"enable_ssh"`
	Diego     bool   `json:"diego"`
	State     string `json:"state"`
}

type CFApp struct {
	Metadata metadata `json:"metadata"`
	Entity   entity   `json:"entity"`
}

func (af *appFactory) Get(appName string) (App, error) {
	output, err := af.cli.CliCommandWithoutTerminalOutput("app", appName, "--guid")
	if err != nil {
		return App{}, errors.New(output[len(output)-1])
	}

	guid := strings.TrimSpace(output[0])

	app := CFApp{}
	err = af.curl(af.cli, &app, "/v2/apps/"+guid)
	if err != nil {
		return App{}, errors.New("Failed to acquire " + appName + " info")
	}

	return App{
		Guid:      app.Metadata.Guid,
		EnableSSH: app.Entity.EnableSSH,
		Diego:     app.Entity.Diego,
		State:     app.Entity.State,
	}, nil
}

func (af *appFactory) SetBool(anApp App, key string, value bool) error {
	return af.curl(af.cli, nil, "/v2/apps/"+anApp.Guid, "-X", "PUT", "-d", `{"`+key+`":`+strconv.FormatBool(value)+`}`)
}
