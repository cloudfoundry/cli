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
// interpolates variables if a vars file is provided, and retunrs a fully
// merged set of applications.
func ReadAndInterpolateManifest(pathToManifest string, pathToVarsFile string) ([]Application, error) {
	rawManifest, err := ioutil.ReadFile(pathToManifest)
	if err != nil {
		return nil, err
	}

	tpl := template.NewTemplate(rawManifest)

	if pathToVarsFile != "" {
		var (
			rawVarsFile []byte
			staticVars  template.StaticVariables
		)

		rawVarsFile, err = ioutil.ReadFile(pathToVarsFile)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(rawVarsFile, &staticVars)
		if err != nil {
			return nil, err
		}

		rawManifest, err = tpl.Evaluate(staticVars, nil, template.EvaluateOpts{})
		if err != nil {
			return nil, err
		}
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
