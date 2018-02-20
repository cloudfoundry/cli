package manifestparser

import (
	"errors"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Application struct {
	Name string `yaml:"name"`
}

type Parser struct {
	PathToManifest string

	Applications []Application

	rawManifest []byte
}

func NewParser() *Parser {
	return new(Parser)
}

func (parser *Parser) Parse(manifestPath string) error {
	bytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	parser.rawManifest = bytes

	var raw struct {
		Applications []Application `yaml:"applications"`
	}

	err = yaml.Unmarshal(bytes, &raw)
	if err != nil {
		return err
	}

	parser.Applications = raw.Applications

	if len(parser.Applications) == 0 {
		return errors.New("must have at least one application")
	}

	return nil
}

func (parser Parser) AppNames() []string {
	var names []string
	for _, app := range parser.Applications {
		names = append(names, app.Name)
	}
	return names
}

func (parser Parser) RawManifest(_ string) ([]byte, error) {
	return parser.rawManifest, nil
}
