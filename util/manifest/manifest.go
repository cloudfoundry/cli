package manifest

import (
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bosh-cli/director/template"
	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Applications []Application `yaml:"applications"`
}

func (manifest *Manifest) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw rawManifest
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	if raw.containsInheritanceField() {
		return InheritanceFieldError{}
	}

	if globals := raw.containsGlobalFields(); len(globals) > 0 {
		return GlobalFieldsError{Fields: globals}
	}

	manifest.Applications = raw.Applications
	return nil
}

// ReadAndInterpolateManifest reads the manifest at the provided paths,
// interpolates variables if a vars file is provided, and returns a fully
// merged set of applications.
func ReadAndInterpolateManifest(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]Application, error) {
	rawManifest, err := ReadAndInterpolateRawManifest(pathToManifest, pathsToVarsFiles, vars)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	err = yaml.Unmarshal(rawManifest, &manifest)
	if err != nil {
		return nil, err
	}

	for i, app := range manifest.Applications {
		if app.Path != "" && !filepath.IsAbs(app.Path) {
			manifest.Applications[i].Path = filepath.Join(filepath.Dir(pathToManifest), app.Path)
		}
	}

	return manifest.Applications, err
}

// ReadAndInterpolateRawManifest reads the manifest at the provided paths,
// interpolates variables if a vars file is provided, and returns the
// Unmarshalled result.
func ReadAndInterpolateRawManifest(pathToManifest string, pathsToVarsFiles []string, vars []template.VarKV) ([]byte, error) {
	rawManifest, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	tpl := template.NewTemplate(rawManifest)
	fileVars := template.StaticVariables{}

	for _, path := range pathsToVarsFiles {
		rawVarsFile, ioerr := ioutil.ReadFile(path)
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

// WriteApplicationManifest writes the provided application to the given
// filepath. If the filepath does not exist, it will create it.
func WriteApplicationManifest(application Application, filePath string) error {
	manifest := Manifest{Applications: []Application{application}}
	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return ManifestCreationError{Err: err}
	}

	err = ioutil.WriteFile(filePath, manifestBytes, 0644)
	if err != nil {
		return ManifestCreationError{Err: err}
	}

	return nil
}
