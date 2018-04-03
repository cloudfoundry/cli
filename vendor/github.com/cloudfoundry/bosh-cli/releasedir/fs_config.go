package releasedir

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

/*
# final.yml
---
name: cf
blobstore:
  provider: s3
  options:
    bucket_name: cf-release-blobs

# private.yml
---
blobstore:
  options: { ... }
*/

type FSConfig struct {
	publicPath  string
	privatePath string
	fs          boshsys.FileSystem
}

type fsConfigPublicSchema struct {
	Name      string                   `yaml:"name"`
	FinalName string                   `yaml:"final_name,omitempty"`
	Blobstore fsConfigSchema_Blobstore `yaml:"blobstore,omitempty"`
}

type fsConfigPrivateSchema struct {
	Blobstore fsConfigSchema_Blobstore `yaml:"blobstore"`
}

type fsConfigSchema_Blobstore struct {
	Provider string                 `yaml:"provider"`
	Options  map[string]interface{} `yaml:"options,omitempty"`
}

func NewFSConfig(publicPath, privatePath string, fs boshsys.FileSystem) FSConfig {
	return FSConfig{publicPath: publicPath, privatePath: privatePath, fs: fs}
}

func (c FSConfig) Name() (string, error) {
	publicSchema, _, err := c.read()
	if err != nil {
		return "", err
	}

	if len(publicSchema.Name) == 0 {
		if len(publicSchema.FinalName) == 0 {
			return "", bosherr.Errorf(
				"Expected non-empty 'name' in config '%s'", c.publicPath)
		}

		return publicSchema.FinalName, nil
	} else if len(publicSchema.FinalName) > 0 {
		return "", bosherr.Errorf(
			"Expected 'name' or 'final_name' but not both in config '%s'", c.publicPath)
	}

	return publicSchema.Name, nil
}

func (c FSConfig) SaveName(name string) error {
	publicSchema, _, err := c.read()
	if err != nil {
		return err
	}

	publicSchema.FinalName = ""
	publicSchema.Name = name

	bytes, err := yaml.Marshal(publicSchema)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling config")
	}

	err = c.fs.WriteFile(c.publicPath, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing config '%s'", c.publicPath)
	}

	return nil
}

func (c FSConfig) Blobstore() (string, map[string]interface{}, error) {
	publicSchema, privateSchema, err := c.read()
	if err != nil {
		return "", nil, err
	}

	if len(publicSchema.Blobstore.Provider) == 0 {
		return "", nil, bosherr.Errorf(
			"Expected non-empty 'blobstore.provider' in config '%s'", c.publicPath)
	}

	opts := map[string]interface{}{}

	for k, v := range publicSchema.Blobstore.Options {
		opts[k] = v
	}

	for k, v := range privateSchema.Blobstore.Options {
		opts[k] = v
	}

	return publicSchema.Blobstore.Provider, opts, nil
}

func (c FSConfig) read() (fsConfigPublicSchema, fsConfigPrivateSchema, error) {
	var publicSchema fsConfigPublicSchema
	var privateSchema fsConfigPrivateSchema

	if c.fs.FileExists(c.publicPath) {
		bytes, err := c.fs.ReadFile(c.publicPath)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Reading config '%s'", c.publicPath)
		}

		err = yaml.Unmarshal(bytes, &publicSchema)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Unmarshalling config '%s'", c.publicPath)
		}
	}

	if c.fs.FileExists(c.privatePath) {
		bytes, err := c.fs.ReadFile(c.privatePath)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Reading config '%s'", c.privatePath)
		}

		err = yaml.Unmarshal(bytes, &privateSchema)
		if err != nil {
			return publicSchema, privateSchema,
				bosherr.WrapErrorf(err, "Unmarshalling config '%s'", c.privatePath)
		}
	}

	return publicSchema, privateSchema, nil
}
