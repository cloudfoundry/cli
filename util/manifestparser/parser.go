package manifestparser

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
)

type Parser struct {
	Applications []Application

	pathToManifest string
	rawManifest    []byte
	validators     []validatorFunc
	hasParsed      bool
}

func NewParser() *Parser {
	return new(Parser)
}

func (parser Parser) AppNames() []string {
	var names []string
	for _, app := range parser.Applications {
		names = append(names, app.Name)
	}
	return names
}

func (parser Parser) Apps() []Application {
	return parser.Applications
}

func (parser Parser) ContainsManifest() bool {
	return parser.hasParsed
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

func (parser Parser) GetPathToManifest() string {
	return parser.pathToManifest
}

// InterpolateAndParse reads the manifest at the provided paths, interpolates
// variables if a vars file is provided, and sets the current manifest to the
// resulting manifest.
// For manifests with only 1 application, appName will override the name of the
// single app defined.
// For manifests with multiple applications, appName will filter the
// applications and leave only a single application in the resulting parsed
// manifest structure.
func (parser *Parser) InterpolateAndParse(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV, appName string) error {
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

	parser.pathToManifest = pathToManifest
	return parser.parse(rawManifest, appName)
}

func (parser Parser) RawAppManifest(appName string) ([]byte, error) {
	var appManifest manifest
	for _, app := range parser.Applications {
		if app.Name == appName {
			appManifest.Applications = []Application{app}
			return yaml.Marshal(appManifest)
		}
	}
	return nil, AppNotInManifestError{Name: appName}
}

func (parser *Parser) parse(manifestBytes []byte, appName string) error {
	parser.rawManifest = manifestBytes
	pathToManifest := parser.GetPathToManifest()
	var raw manifest

	err := yaml.Unmarshal(manifestBytes, &raw)
	if err != nil {
		return err
	}

	if len(raw.Applications) == 0 {
		return errors.New("must have at least one application")
	}

	if len(raw.Applications) == 1 && appName != "" {
		raw.Applications[0].Name = appName
		raw.Applications[0].FullUnmarshalledApplication["name"] = appName
	}

	filteredIndex := -1

	for i := range raw.Applications {
		if raw.Applications[i].Name == "" {
			return errors.New("Found an application with no name specified")
		}

		if raw.Applications[i].Name == appName {
			filteredIndex = i
			break
		}

		if raw.Applications[i].Path == "" {
			continue
		}

		var finalPath = raw.Applications[i].Path
		if !filepath.IsAbs(finalPath) {
			finalPath = filepath.Join(filepath.Dir(pathToManifest), finalPath)
		}
		finalPath, err = filepath.EvalSymlinks(finalPath)
		if err != nil {
			if os.IsNotExist(err) {
				return InvalidManifestApplicationPathError{
					Path: raw.Applications[i].Path,
				}
			}
			return err
		}
		raw.Applications[i].Path = finalPath
	}

	if filteredIndex >= 0 {
		raw.Applications = []Application{raw.Applications[filteredIndex]}
	} else if appName != "" {
		return AppNotInManifestError{Name: appName}
	}

	parser.Applications = raw.Applications
	parser.rawManifest, err = yaml.Marshal(raw)
	if err != nil {
		return err
	}

	parser.hasParsed = true
	return nil
}
