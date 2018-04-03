package releasedir

import (
	"fmt"
	"path/filepath"
	"sort"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	semver "github.com/cppforlife/go-semi-semantic/version"
	"gopkg.in/yaml.v2"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
	boshrelman "github.com/cloudfoundry/bosh-cli/release/manifest"
)

type FSReleaseIndex struct {
	name     string
	dirPath  string
	reporter ReleaseIndexReporter

	uuidGen boshuuid.Generator
	fs      boshsys.FileSystem
}

/*
---
builds:
  70b9ea8efb83b882021792517b1164550b41bc27:
    version: "1"
format-version: "2"
*/

type fsReleaseIndexSchema struct {
	Builds fsReleaseIndexSchema_SortedEntries `yaml:"builds"`

	FormatVersion string `yaml:"format-version"`
}

type fsReleaseIndexSchema_SortedEntries map[string]fsReleaseIndexSchema_Entry

var _ yaml.Marshaler = fsReleaseIndexSchema_SortedEntries{}

func (e fsReleaseIndexSchema_SortedEntries) MarshalYAML() (interface{}, error) {
	var keys []string
	for k, _ := range e {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))
	var sortedEntries []yaml.MapItem
	for _, k := range keys {
		sortedEntries = append(sortedEntries, yaml.MapItem{Key: k, Value: e[k]})
	}
	return sortedEntries, nil
}

type fsReleaseIndexSchema_Entry struct {
	Version string `yaml:"version"`
}

type releaseIndexEntry struct {
	Version semver.Version
}

func NewFSReleaseIndex(
	name string,
	dirPath string,
	reporter ReleaseIndexReporter,
	uuidGen boshuuid.Generator,
	fs boshsys.FileSystem,
) FSReleaseIndex {
	return FSReleaseIndex{
		name:     name,
		dirPath:  dirPath,
		reporter: reporter,
		uuidGen:  uuidGen,
		fs:       fs,
	}
}

func (i FSReleaseIndex) LastVersion(name string) (*semver.Version, error) {
	if len(name) == 0 {
		return nil, bosherr.Error("Expected non-empty release name")
	}

	schema, err := i.read(name)
	if err != nil {
		return nil, err
	}

	var versions []semver.Version

	for _, entry := range schema.Builds {
		ver, err := semver.NewVersionFromString(entry.Version)
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Parsing release versions")
		}

		versions = append(versions, ver)
	}

	if len(versions) == 0 {
		return nil, nil
	}

	sort.Sort(semver.AscSorting(versions))

	return &versions[len(versions)-1], nil
}

func (i FSReleaseIndex) Contains(release boshrel.Release) (bool, error) {
	if len(release.Name()) == 0 {
		return false, bosherr.Error("Expected non-empty release name")
	}

	if len(release.Version()) == 0 {
		return false, bosherr.Error("Expected non-empty release version")
	}

	schema, err := i.read(release.Name())
	if err != nil {
		return false, err
	}

	for _, entry := range schema.Builds {
		if entry.Version == release.Version() {
			return true, nil
		}
	}

	return false, nil
}

func (i FSReleaseIndex) Add(manifest boshrelman.Manifest) error {
	if len(manifest.Name) == 0 {
		return bosherr.Error("Expected non-empty release name")
	}

	if len(manifest.Version) == 0 {
		return bosherr.Error("Expected non-empty release version")
	}

	schema, err := i.read(manifest.Name)
	if err != nil {
		return err
	}

	for _, entry := range schema.Builds {
		if entry.Version == manifest.Version {
			return bosherr.Errorf("Release version '%s' already exists", manifest.Version)
		}
	}

	uuid, err := i.uuidGen.Generate()
	if err != nil {
		return bosherr.WrapErrorf(err, "Generating key for release index entry")
	}

	schema.Builds[uuid] = fsReleaseIndexSchema_Entry{Version: manifest.Version}

	desc := fmt.Sprintf("%s/%s", manifest.Name, manifest.Version)

	err = i.saveManifest(manifest)
	if err != nil {
		i.reporter.ReleaseIndexAdded(i.name, desc, err)
		return err
	}

	err = i.save(manifest.Name, schema)
	if err != nil {
		i.reporter.ReleaseIndexAdded(i.name, desc, err)
		return err
	}

	i.reporter.ReleaseIndexAdded(i.name, desc, nil)

	return nil
}

func (i FSReleaseIndex) ManifestPath(name, version string) string {
	fileName := fmt.Sprintf("%s-%s.yml", name, version)

	return filepath.Join(i.dirPath, name, fileName)
}

func (i FSReleaseIndex) indexPath(name string) string {
	return filepath.Join(i.dirPath, name, "index.yml")
}

func (i FSReleaseIndex) read(name string) (fsReleaseIndexSchema, error) {
	var schema fsReleaseIndexSchema

	// Default to an empty map
	schema.Builds = fsReleaseIndexSchema_SortedEntries{}

	indexPath := i.indexPath(name)

	if !i.fs.FileExists(indexPath) {
		return schema, nil
	}

	bytes, err := i.fs.ReadFile(indexPath)
	if err != nil {
		return schema, bosherr.WrapErrorf(err, "Reading index")
	}

	err = yaml.Unmarshal(bytes, &schema)
	if err != nil {
		return schema, bosherr.WrapError(err, "Unmarshalling index")
	}

	return schema, nil
}

func (i FSReleaseIndex) save(name string, schema fsReleaseIndexSchema) error {
	schema.FormatVersion = "2"

	bytes, err := yaml.Marshal(schema)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling index")
	}

	indexPath := i.indexPath(name)

	err = i.fs.WriteFile(indexPath, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing index")
	}

	return nil
}

func (i FSReleaseIndex) saveManifest(manifest boshrelman.Manifest) error {
	bytes, err := yaml.Marshal(manifest)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling manifest")
	}

	manifestPath := i.ManifestPath(manifest.Name, manifest.Version)

	err = i.fs.WriteFile(manifestPath, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing manifest")
	}

	return nil
}
