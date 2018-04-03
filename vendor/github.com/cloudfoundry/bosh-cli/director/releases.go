package director

import (
	"encoding/json"
	"fmt"
	"net/http"
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"
)

type ReleaseImpl struct {
	client Client

	name    string
	version semver.Version

	currentlyDeployed bool

	commitHash         string
	uncommittedChanges bool

	jobs     []Job
	packages []Package

	fetched  bool
	fetchErr error
}

func (r ReleaseImpl) Name() string            { return r.name }
func (r ReleaseImpl) Version() semver.Version { return r.version }

func (r ReleaseImpl) VersionMark(suffix string) string {
	if r.currentlyDeployed {
		return suffix
	}
	return ""
}

func (r ReleaseImpl) CommitHashWithMark(suffix string) string {
	if r.uncommittedChanges {
		return r.commitHash + suffix
	}
	return r.commitHash
}

func (r *ReleaseImpl) Jobs() ([]Job, error) {
	r.fetch()
	return r.jobs, r.fetchErr
}

func (r *ReleaseImpl) Packages() ([]Package, error) {
	r.fetch()
	return r.packages, r.fetchErr
}

func (r *ReleaseImpl) fetch() {
	if r.fetched {
		return
	}

	resp, err := r.client.Release(r.name, r.version.String())
	if err != nil {
		r.fetchErr = bosherr.WrapErrorf(err, "Expected to find release '%s/%s'", r.name, r.version)
		return
	}

	r.fill(resp)
}

func (r *ReleaseImpl) fill(resp ReleaseResp) {
	r.fetched = true

	r.jobs = resp.Jobs
	r.packages = resp.Packages
}

func (r ReleaseImpl) Delete(force bool) error {
	err := r.client.DeleteReleaseOrSeries(r.name, r.version.String(), force)
	if err != nil {
		found, listErr := r.client.HasRelease(r.name, r.version.String())
		if found || listErr != nil {
			return err
		}
	}

	return nil
}

type ReleaseResp struct {
	Jobs     []Job
	Packages []Package
}

type Job struct {
	Name        string
	Fingerprint string

	BlobstoreID string `json:"blobstore_id"`
	SHA1        string `json:"sha1"`

	LinksConsumed []Link `json:"consumes"`
	LinksProvided []Link `json:"provides"`
}

type Link struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Optional bool   `json:"optional" yaml:",omitempty"`
}

type Package struct {
	Name        string
	Fingerprint string

	BlobstoreID string `json:"blobstore_id"`
	SHA1        string `json:"sha1"`

	CompiledPackages []CompiledPackage `json:"compiled_packages"`
}

type CompiledPackage struct {
	Stemcell OSVersionSlug `json:"stemcell"` // e.g. "ubuntu-trusty/3093"

	BlobstoreID string `json:"blobstore_id"`
	SHA1        string `json:"sha1"`
}

func (d DirectorImpl) Releases() ([]Release, error) {
	var rels []Release

	resps, err := d.client.ReleaseSeries()
	if err != nil {
		return rels, err
	}

	for _, resp := range resps {
		for _, relVerResp := range resp.Versions {
			parsedVersion, err := semver.NewVersionFromString(relVerResp.Version)
			if err != nil {
				return nil, bosherr.WrapErrorf(
					err, "Parsing version for release '%s/%s'", resp.Name, relVerResp.Version)
			}

			rel := &ReleaseImpl{
				client: d.client,

				name:    resp.Name,
				version: parsedVersion,

				currentlyDeployed: relVerResp.CurrentlyDeployed,

				commitHash:         relVerResp.CommitHash,
				uncommittedChanges: relVerResp.UncommittedChanges,
			}

			rels = append(rels, rel)
		}
	}

	return rels, nil
}

func (d DirectorImpl) FindRelease(slug ReleaseSlug) (Release, error) {
	parsedVersion, err := semver.NewVersionFromString(slug.Version())
	if err != nil {
		return nil, bosherr.WrapErrorf(
			err, "Parsing version for release '%s/%s'", slug.Name(), slug.Version())
	}

	rel := &ReleaseImpl{
		client:  d.client,
		name:    slug.Name(),
		version: parsedVersion,
	}

	return rel, nil
}

func (d DirectorImpl) HasRelease(name, version string, stemcell OSVersionSlug) (bool, error) {
	found, err := d.client.HasRelease(name, version)
	if err != nil {
		return false, err
	}

	if !stemcell.IsProvided() || !found {
		return found, nil
	}

	return d.releaseHasCompiledPackage(NewReleaseSlug(name, version), stemcell)
}

// releaseHasCompiledPackage returns true if release contains
// at least one compiled package that matches stemcell slug
func (d DirectorImpl) releaseHasCompiledPackage(releaseSlug ReleaseSlug, osVersionSlug OSVersionSlug) (bool, error) {
	release, err := d.FindRelease(releaseSlug)
	if err != nil {
		return false, err
	}

	pkgs, err := release.Packages()
	if err != nil {
		return false, err
	}

	for _, pkg := range pkgs {
		for _, compiledPkg := range pkg.CompiledPackages {
			if compiledPkg.Stemcell == osVersionSlug {
				return true, nil
			}
		}
	}

	return false, nil
}

func (d DirectorImpl) UploadReleaseURL(url, sha1 string, rebase, fix bool) error {
	return d.client.UploadReleaseURL(url, sha1, rebase, fix)
}

func (d DirectorImpl) UploadReleaseFile(file UploadFile, rebase, fix bool) error {
	return d.client.UploadReleaseFile(file, rebase, fix)
}

func (c Client) Release(name, version string) (ReleaseResp, error) {
	var resp ReleaseResp

	if len(name) == 0 {
		return resp, bosherr.Error("Expected non-empty release name")
	}

	if len(version) == 0 {
		return resp, bosherr.Error("Expected non-empty release version")
	}

	query := gourl.Values{}

	query.Add("version", version)

	path := fmt.Sprintf("/releases/%s?%s", name, query.Encode())

	err := c.clientRequest.Get(path, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Finding release '%s/%s'", name, version)
	}

	return resp, nil
}

func (c Client) HasRelease(name, version string) (bool, error) {
	resps, err := c.ReleaseSeries()
	if err != nil {
		return false, err
	}

	for _, r := range resps {
		if r.Name == name {
			for _, v := range r.Versions {
				if v.Version == version {
					return true, nil
				}
			}
			return false, nil
		}
	}

	return false, nil
}

func (c Client) UploadReleaseURL(url, sha1 string, rebase, fix bool) error {
	if len(url) == 0 {
		return bosherr.Error("Expected non-empty URL")
	}

	query := gourl.Values{}

	if rebase {
		query.Add("rebase", "true")
	}

	if fix {
		query.Add("fix", "true")
	}

	path := "/releases?" + query.Encode()

	body := map[string]string{"location": url}

	if len(sha1) > 0 {
		body["sha1"] = sha1
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, err = c.taskClientRequest.PostResult(path, reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Uploading remote release '%s'", url)
	}

	return nil
}

func (c Client) UploadReleaseFile(file UploadFile, rebase, fix bool) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return bosherr.WrapErrorf(err, "Determining release file size")
	}

	query := gourl.Values{}

	if rebase {
		query.Add("rebase", "true")
	}

	if fix {
		query.Add("fix", "true")
	}

	path := "/releases?" + query.Encode()

	setHeadersAndBody := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/x-compressed")
		req.ContentLength = fileInfo.Size()
		req.Body = file
	}

	_, err = c.taskClientRequest.PostResult(path, nil, setHeadersAndBody)
	if err != nil {
		return bosherr.WrapErrorf(err, "Uploading release file")
	}

	return nil
}
