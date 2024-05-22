package manifestparser

import (
	"errors"
	"os"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"gopkg.in/yaml.v2"
)

type ManifestParser struct{}

// InterpolateAndParse reads the manifest at the provided paths, interpolates
// variables if a vars file is provided, and sets the current manifest to the
// resulting manifest.
// For manifests with only 1 application, appName will override the name of the
// single app defined.
// For manifests with multiple applications, appName will filter the
// applications and leave only a single application in the resulting parsed
// manifest structure.
func (m ManifestParser) InterpolateManifest(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]byte, error) {
	rawManifest, err := os.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	tpl := template.NewTemplate(rawManifest)
	fileVars := template.StaticVariables{}

	for _, path := range pathsToVarsFiles {
		rawVarsFile, ioerr := os.ReadFile(path)
		if ioerr != nil {
			return nil, ioerr
		}

		var sv template.StaticVariables

		err = yaml.Unmarshal(rawVarsFile, &sv)
		if err != nil {
			return nil, InvalidYAMLError{Err: err}
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
		return nil, InterpolationError{Err: err}
	}

	return rawManifest, nil
}

func (m ManifestParser) ParseManifest(pathToManifest string, rawManifest []byte) (Manifest, error) {
	var parsedManifest Manifest
	err := yaml.Unmarshal(rawManifest, &parsedManifest)
	if err != nil {
		return Manifest{}, &yaml.TypeError{}
	}

	if len(parsedManifest.Applications) == 0 {
		return Manifest{}, errors.New("Manifest must have at least one application.")
	}

	parsedManifest.PathToManifest = pathToManifest

	return parsedManifest, nil
}

func (m ManifestParser) MarshalManifest(manifest Manifest) ([]byte, error) {
	return yaml.Marshal(manifest)
}
