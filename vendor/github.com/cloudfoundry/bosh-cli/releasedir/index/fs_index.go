package index

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type FSIndex struct {
	name     string
	dirPath  string
	reporter Reporter
	blobs    IndexBlobs
	fs       boshsys.FileSystem

	useSubdir           bool
	expectsBlobstoreIDs bool
	mutex               *sync.Mutex
}

type indexEntry struct {
	Key     string
	Version string

	BlobstoreID string
	SHA1        string
}

/*
---
builds:
  70b9ea8efb83b882021792517b1164550b41bc27:
    version: 70b9ea8efb83b882021792517b1164550b41bc27
    sha1: b514a8cc194278d5c5b96d552cdbe0cade1ed719
format-version: '2'
---
builds:
  88a17b61a5b892c9c59884ea5dee04ba6e116bb2:
    version: 88a17b61a5b892c9c59884ea5dee04ba6e116bb2
    sha1: 98eca647ddf3b24ae62fce44ea8a8e3526c3e53c
    blobstore_id: cd26953d-b811-4127-b512-6f48639bd069
format-version: '2'
*/

type duplicateError struct {
	msg  string
	args []interface{}
}

func (de duplicateError) Error() string {
	return fmt.Sprintf(de.msg, de.args...)
}

func (de duplicateError) IsDuplicate() bool {
	return true
}

type fsIndexSchema struct {
	Builds fsIndexSchema_SortedEntries `yaml:"builds"`

	FormatVersion string `yaml:"format-version"`
}

type fsIndexSchema_SortedEntries map[string]fsIndexSchema_Entry

var _ yaml.Marshaler = fsIndexSchema_SortedEntries{}

func (e fsIndexSchema_SortedEntries) MarshalYAML() (interface{}, error) {
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

type fsIndexSchema_Entry struct {
	Version string `yaml:"version"`

	BlobstoreID string `yaml:"blobstore_id,omitempty"`
	SHA1        string `yaml:"sha1"`
}

func NewFSIndex(
	name string,
	dirPath string,
	useSubdir bool,
	expectsBlobstoreIDs bool,
	reporter Reporter,
	blobs IndexBlobs,
	fs boshsys.FileSystem,
) FSIndex {
	return FSIndex{
		name:     name,
		dirPath:  dirPath,
		reporter: reporter,
		blobs:    blobs,
		fs:       fs,

		useSubdir:           useSubdir,
		expectsBlobstoreIDs: expectsBlobstoreIDs,
		mutex:               &sync.Mutex{},
	}
}

func (i FSIndex) Find(name, fingerprint string) (string, string, error) {
	if len(name) == 0 {
		return "", "", bosherr.Error("Expected non-empty name")
	}

	if len(fingerprint) == 0 {
		return "", "", bosherr.Error("Expected non-empty fingerprint")
	}

	entries, err := i.entries(name)
	if err != nil {
		return "", "", err
	}

	for _, entry := range entries {
		if entry.Version == fingerprint {
			blobName := fmt.Sprintf("%s/%s", name, entry.Version)
			blobPath, err := i.blobs.Get(blobName, entry.BlobstoreID, entry.SHA1)
			if err != nil {
				return "", "", err
			}

			return blobPath, entry.SHA1, nil
		}
	}

	return "", "", nil
}

func (i FSIndex) Add(name, fingerprint, path, sha1 string) (string, string, error) {
	if len(name) == 0 {
		return "", "", bosherr.Error("Expected non-empty name")
	}

	if len(fingerprint) == 0 {
		return "", "", bosherr.Error("Expected non-empty fingerprint")
	}

	if len(path) == 0 {
		return "", "", bosherr.Error("Expected non-empty archive path")
	}

	if len(sha1) == 0 {
		return "", "", bosherr.Error("Expected non-empty archive SHA1")
	}

	entries, err := i.entries(name)
	if err != nil {
		return "", "", err
	}

	for _, entry := range entries {
		if entry.Version == fingerprint {
			return "", "", duplicateError{
				msg:  "Trying to add duplicate index entry '%s/%s' and SHA1 '%s' (conflicts with '%#v')",
				args: []interface{}{name, fingerprint, sha1, entry},
			}
		}
	}

	entry := indexEntry{
		Key:     fingerprint,
		Version: fingerprint,
		SHA1:    sha1,
	}

	desc := fmt.Sprintf("%s/%s", name, fingerprint)

	i.reporter.IndexEntryStartedAdding(i.name, desc)

	blobID, blobPath, err := i.blobs.Add(desc, path, sha1)
	if err != nil {
		i.reporter.IndexEntryFinishedAdding(i.name, desc, err)
		return "", "", err
	}

	entry.BlobstoreID = blobID

	entries = append(entries, entry)

	err = i.save(name, entries)
	if err != nil {
		i.reporter.IndexEntryFinishedAdding(i.name, desc, err)
		return "", "", err
	}

	i.reporter.IndexEntryFinishedAdding(i.name, desc, nil)

	return blobPath, sha1, nil
}

var (
	// Ruby CLI for some reason produces invalid annotations
	invalidBinaryAnnotationReplacer = strings.NewReplacer(" !binary ", " !!binary ")
)

func (i FSIndex) entries(name string) ([]indexEntry, error) {
	indexPath := i.indexPath(name)

	if !i.fs.FileExists(indexPath) {
		return nil, nil
	}

	bytes, err := i.fs.ReadFile(indexPath)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading index")
	}

	var schema fsIndexSchema

	str := invalidBinaryAnnotationReplacer.Replace(string(bytes))

	err = yaml.Unmarshal([]byte(str), &schema)
	if err != nil {
		return nil, bosherr.WrapError(err, "Unmarshalling index")
	}

	var entries []indexEntry

	for versionKey, entry := range schema.Builds {
		entries = append(entries, indexEntry{
			Key:     versionKey,
			Version: entry.Version,

			BlobstoreID: entry.BlobstoreID,
			SHA1:        entry.SHA1,
		})
	}

	return entries, nil
}

func (i FSIndex) save(name string, entries []indexEntry) error {
	schema := fsIndexSchema{
		Builds:        fsIndexSchema_SortedEntries{},
		FormatVersion: "2",
	}

	for _, entry := range entries {
		if i.expectsBlobstoreIDs && len(entry.BlobstoreID) == 0 {
			return bosherr.Errorf("Internal inconsistency: entry must include blob ID '%#v'", entry)
		} else if !i.expectsBlobstoreIDs && len(entry.BlobstoreID) > 0 {
			return bosherr.Errorf("Internal inconsistency: entry must not include blob ID '%#v'", entry)
		}

		schema.Builds[entry.Key] = fsIndexSchema_Entry{
			Version:     entry.Version,
			BlobstoreID: entry.BlobstoreID,
			SHA1:        entry.SHA1,
		}
	}

	bytes, err := yaml.Marshal(schema)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling index")
	}

	indexPath := i.indexPath(name)

	i.mutex.Lock()
	defer i.mutex.Unlock()

	err = i.fs.WriteFile(indexPath, bytes)
	if err != nil {
		return bosherr.WrapErrorf(err, "Writing index")
	}

	return nil
}

func (i FSIndex) indexPath(name string) string {
	if i.useSubdir {
		return filepath.Join(i.dirPath, name, "index.yml")
	}
	return filepath.Join(i.dirPath, "index.yml")
}
