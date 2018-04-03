package director

import (
	"fmt"
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ReleaseSeriesImpl struct {
	client Client

	name string
}

func (rs ReleaseSeriesImpl) Name() string { return rs.name }

func (rs ReleaseSeriesImpl) Delete(force bool) error {
	err := rs.client.DeleteReleaseOrSeries(rs.name, "", force)
	if err != nil {
		resps, listErr := rs.client.ReleaseSeries()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.Name == rs.name {
				return err
			}
		}
	}

	return nil
}

type ReleaseSeriesResp struct {
	Name     string
	Versions []ReleaseVersionResp `json:"release_versions"`
}

type ReleaseVersionResp struct {
	Version string

	CurrentlyDeployed bool `json:"currently_deployed"`

	CommitHash         string `json:"commit_hash"`
	UncommittedChanges bool   `json:"uncommitted_changes"`
}

func (d DirectorImpl) FindReleaseSeries(slug ReleaseSeriesSlug) (ReleaseSeries, error) {
	return ReleaseSeriesImpl{client: d.client, name: slug.Name()}, nil
}

func (c Client) ReleaseSeries() ([]ReleaseSeriesResp, error) {
	var resps []ReleaseSeriesResp

	err := c.clientRequest.Get("/releases", &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Finding releases")
	}

	return resps, nil
}

func (c Client) DeleteReleaseOrSeries(name, version string, force bool) error {
	if len(name) == 0 {
		return bosherr.Error("Expected non-empty release name")
	}

	// version may be empty

	query := gourl.Values{}

	if len(version) > 0 {
		query.Add("version", version)
	}

	if force {
		query.Add("force", "true")
	}

	path := fmt.Sprintf("/releases/%s?%s", name, query.Encode())

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting release or series '%s[/%s]'", name, version)
	}

	return nil
}
