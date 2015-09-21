package space

import (
	"errors"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/models"
	"github.com/cloudfoundry/cli/plugin"
)

//go:generate counterfeiter -o space_fakes/fake_space_factory.go . SpaceFactory
type SpaceFactory interface {
	Get(string) (Space, error)
	SetBool(aSpace Space, key string, value bool) error
}

type spaceFactory struct {
	cli  plugin.CliConnection
	curl models.Curler
}

func NewSpaceFactory(cli plugin.CliConnection, curl models.Curler) SpaceFactory {
	return &spaceFactory{cli: cli, curl: curl}
}

type Space struct {
	Guid     string
	AllowSSH bool
}

type metadata struct {
	Guid string `json:"guid"`
}

type entity struct {
	AllowSSH bool `json:"allow_ssh"`
}

type CFSpace struct {
	Metadata metadata `json:"metadata"`
	Entity   entity   `json:"entity"`
}

func (sf *spaceFactory) Get(spaceName string) (Space, error) {
	output, err := sf.cli.CliCommandWithoutTerminalOutput("space", spaceName, "--guid")
	if err != nil {
		return Space{}, errors.New(output[len(output)-1])
	}

	guid := strings.TrimSpace(output[0])
	space := CFSpace{}
	err = sf.curl(sf.cli, &space, "/v2/spaces/"+guid)

	if err != nil {
		return Space{}, errors.New("Failed to acquire " + spaceName + " info")
	}

	return Space{
		Guid:     space.Metadata.Guid,
		AllowSSH: space.Entity.AllowSSH,
	}, nil
}

func (sf *spaceFactory) SetBool(aSpace Space, key string, value bool) error {
	return sf.curl(sf.cli, nil, "/v2/spaces/"+aSpace.Guid, "-X", "PUT", "-d", `{"`+key+`":`+strconv.FormatBool(value)+`}`)
}
