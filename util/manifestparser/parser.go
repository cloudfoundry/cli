package manifestparser

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
)

type Parser struct {
	PathToManifest string

	Applications []Application

	rawManifest []byte

	validators []validatorFunc
}

func NewParser() *Parser {
	parser := new(Parser)
	parser.validators = []validatorFunc{
		ValidatePaths,
	}
	return parser
}

func (parser *Parser) Parse(manifestPath string) error {
	bytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	return parser.parse(bytes)
}

func (parser *Parser) GetPathToManifest() string {
	return parser.PathToManifest
}

func (parser *Parser) Apps(appName string) ([]Application, error) {
	if appName == "" {
		return parser.Applications, nil
	}
	for _, app := range parser.Applications {
		if app.Name == appName {
			return []Application{app}, nil
		}
	}
	return nil, AppNotInManifestError{Name: appName}
}

// InterpolateAndParse reads the manifest at the provided paths, interpolates
// variables if a vars file is provided, and sets the current manifest to the
// resulting manifest.
func (parser *Parser) InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) error {
	rawManifest, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return err
	}

	tpl := template.NewTemplate(rawManifest)
	fileVars := template.StaticVariables{}

	for _, path := range pathsToVarsFiles {
		rawVarsFile, ioerr := ioutil.ReadFile(path)
		if ioerr != nil {
			return ioerr
		}

		var sv template.StaticVariables

		err = yaml.Unmarshal(rawVarsFile, &sv)
		if err != nil {
			return InvalidYAMLError{Err: err}
		}

		for k, v := range sv {
			fileVars[k] = v
		}
	}

	for _, kv := range vars {
		fileVars[kv.Name] = kv.Value
	}

	rawManifest, err = tpl.Evaluate(fileVars, nil, template.EvaluateOpts{ExpectAllKeys: true})
	if err != nil {
		return InterpolationError{Err: err}
	}

	parser.PathToManifest = pathToManifest
	return parser.parse(rawManifest)
}

func (parser Parser) AppNames() []string {
	var names []string
	for _, app := range parser.Applications {
		names = append(names, app.Name)
	}
	return names
}

func (parser Parser) ContainsMultipleApps() bool {
	return len(parser.Applications) > 1
}

func (parser Parser) ContainsPrivateDockerImages() bool {
	for _, app := range parser.Applications {
		if app.Docker != nil && app.Docker.Username != "" {
			return true
		}
	}
	return false
}

func (parser Parser) FullRawManifest() []byte {
	return parser.rawManifest
}

func (parser Parser) RawAppManifest(appName string) ([]byte, error) {
	var appManifest rawManifest
	for _, app := range parser.Applications {
		if app.Name == appName {
			appManifest.Applications = []Application{app}
			return yaml.Marshal(appManifest)
		}
	}
	return nil, AppNotInManifestError{Name: appName}
}

type rawManifest struct {
	Applications []Application `yaml:"applications"`
}

func (parser *Parser) parse(manifestBytes []byte) error {
	parser.rawManifest = manifestBytes
	pathToManifest := parser.GetPathToManifest()
	var raw rawManifest

	err := yaml.Unmarshal(manifestBytes, &raw)
	if err != nil {
		return err
	}

	parser.Applications = raw.Applications

	if len(parser.Applications) == 0 {
		return errors.New("must have at least one application")
	}

	for index, application := range parser.Applications {
		if application.Name == "" {
			return errors.New("Found an application with no name specified")
		}

		if application.Path == "" {
			continue
		}

		var finalPath = application.Path
		if !filepath.IsAbs(finalPath) {
			finalPath = filepath.Join(filepath.Dir(pathToManifest), finalPath)
		}
		finalPath, err = filepath.EvalSymlinks(finalPath)
		if err != nil {
			return err
		}
		parser.Applications[index].Path = finalPath

	}
	return nil
}

//go:generate counterfeiter . ManifestParser

type ManifestParser interface {
	Apps(appName string) ([]Application, error)
	Validate() error
	ContainsMultipleApps() bool
	InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) error
	FullRawManifest() []byte
	AppNames() []string
	RawAppManifest(appName string) ([]byte, error)
	GetPathToManifest() string
}
