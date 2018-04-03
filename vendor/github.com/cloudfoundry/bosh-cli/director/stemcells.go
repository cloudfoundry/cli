package director

import (
	"encoding/json"
	"fmt"
	"net/http"
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"
)

type StemcellImpl struct {
	client Client

	name              string
	version           semver.Version
	currentlyDeployed bool

	osName string

	cpi string
	cid string
}

func (s StemcellImpl) Name() string            { return s.name }
func (s StemcellImpl) Version() semver.Version { return s.version }

func (s StemcellImpl) OSName() string { return s.osName }
func (s StemcellImpl) CID() string    { return s.cid }
func (s StemcellImpl) CPI() string    { return s.cpi }

func (s StemcellImpl) VersionMark(suffix string) string {
	if s.currentlyDeployed {
		return suffix
	}
	return ""
}

func (s StemcellImpl) Delete(force bool) error {
	err := s.client.DeleteStemcell(s.name, s.version.String(), force)
	if err != nil {
		found, listErr := s.client.HasStemcell(s.name, s.version.String())
		if found || listErr != nil {
			return err
		}
	}

	return nil
}

type StemcellResp struct {
	Name    string
	Version string

	OperatingSystem string `json:"operating_system"`

	CID string `json:"cid"`
	CPI string `json:"cpi"`

	// Only used for determining if stemcell is deployed
	Deployments []interface{}
}

type StemcellInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (d DirectorImpl) Stemcells() ([]Stemcell, error) {
	resps, err := d.client.Stemcells()
	if err != nil {
		return nil, err
	}

	var stems []Stemcell

	for _, resp := range resps {
		parsedVersion, err := semver.NewVersionFromString(resp.Version)
		if err != nil {
			return nil, bosherr.WrapErrorf(
				err, "Parsing version for stemcell '%s/%s'", resp.Name, resp.Version)
		}

		stem := StemcellImpl{
			client: d.client,

			name:              resp.Name,
			version:           parsedVersion,
			currentlyDeployed: len(resp.Deployments) > 0,

			osName: resp.OperatingSystem,

			cpi: resp.CPI,
			cid: resp.CID,
		}

		stems = append(stems, stem)
	}

	return stems, nil
}

func (d DirectorImpl) FindStemcell(slug StemcellSlug) (Stemcell, error) {
	parsedVersion, err := semver.NewVersionFromString(slug.Version())
	if err != nil {
		return nil, bosherr.WrapErrorf(
			err, "Parsing version for stemcell '%s/%s'", slug.Name(), slug.Version())
	}

	stem := StemcellImpl{
		client:  d.client,
		name:    slug.Name(),
		version: parsedVersion,
	}

	return stem, nil
}

func (d DirectorImpl) HasStemcell(name, version string) (bool, error) {
	return d.client.HasStemcell(name, version)
}

func (d DirectorImpl) StemcellNeedsUpload(stemcells StemcellInfo) (bool, bool, error) {
	return d.client.StemcellNeedsUpload(stemcells)
}

func (d DirectorImpl) UploadStemcellURL(url, sha1 string, fix bool) error {
	return d.client.UploadStemcellURL(url, sha1, fix)
}

func (d DirectorImpl) UploadStemcellFile(file UploadFile, fix bool) error {
	return d.client.UploadStemcellFile(file, fix)
}

func (c Client) Stemcells() ([]StemcellResp, error) {
	var resps []StemcellResp

	err := c.clientRequest.Get("/stemcells", &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Finding stemcells")
	}

	return resps, nil
}

func (c Client) HasStemcell(name, version string) (bool, error) {
	resps, err := c.Stemcells()
	if err != nil {
		return false, err
	}

	for _, r := range resps {
		if r.Name == name && r.Version == version {
			return true, nil
		}
	}

	return false, nil
}

func (c Client) StemcellNeedsUpload(stemcells StemcellInfo) (bool, bool, error) {
	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	jsonBody, err := json.Marshal(map[string]StemcellInfo{"stemcell": stemcells})
	if err != nil {
		return false, true, err
	}

	respBody, response, err := c.clientRequest.RawPost("/stemcell_uploads", jsonBody, setHeaders)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			return false, false, bosherr.WrapErrorf(err, "Finding stemcells")
		}
		return false, true, bosherr.WrapErrorf(err, "Finding stemcells")
	}

	var parsedResponse struct {
		Needed bool
	}

	err = json.Unmarshal(respBody, &parsedResponse)
	if err != nil {
		return false, true, bosherr.WrapError(err, "Unmarshaling stemcell matches")
	}

	return parsedResponse.Needed, true, nil
}

func (c Client) UploadStemcellURL(url, sha1 string, fix bool) error {
	if len(url) == 0 {
		return bosherr.Error("Expected non-empty URL")
	}

	query := gourl.Values{}

	if fix {
		query.Add("fix", "true")
	}

	path := "/stemcells?" + query.Encode()

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
		return bosherr.WrapErrorf(err, "Uploading remote stemcell '%s'", url)
	}

	return nil
}

func (c Client) UploadStemcellFile(file UploadFile, fix bool) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return bosherr.WrapErrorf(err, "Determining stemcell file size")
	}

	query := gourl.Values{}

	if fix {
		query.Add("fix", "true")
	}

	path := "/stemcells?" + query.Encode()

	setHeadersAndBody := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/x-compressed")
		req.ContentLength = fileInfo.Size()
		req.Body = file
	}

	_, err = c.taskClientRequest.PostResult(path, nil, setHeadersAndBody)
	if err != nil {
		return bosherr.WrapErrorf(err, "Uploading stemcell file")
	}

	return nil
}

func (c Client) DeleteStemcell(name, version string, force bool) error {
	query := gourl.Values{}

	if force {
		query.Add("force", "true")
	}

	path := fmt.Sprintf("/stemcells/%s/%s?%s", name, version, query.Encode())

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting stemcell '%s/%s'", name, version)
	}

	return nil
}
