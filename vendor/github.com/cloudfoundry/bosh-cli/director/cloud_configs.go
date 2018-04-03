package director

import (
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CloudConfig struct {
	Properties string
}

func (d DirectorImpl) LatestCloudConfig() (CloudConfig, error) {
	resps, err := d.client.CloudConfigs()
	if err != nil {
		return CloudConfig{}, err
	}

	if len(resps) == 0 {
		return CloudConfig{}, bosherr.Error("No cloud config")
	}

	return resps[0], nil
}

func (d DirectorImpl) UpdateCloudConfig(manifest []byte) error {
	return d.client.UpdateCloudConfig(manifest)
}

func (c Client) CloudConfigs() ([]CloudConfig, error) {
	var resps []CloudConfig

	err := c.clientRequest.Get("/cloud_configs?limit=1", &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Finding cloud configs")
	}

	return resps, nil
}

func (c Client) UpdateCloudConfig(manifest []byte) error {
	path := "/cloud_configs"

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	_, _, err := c.clientRequest.RawPost(path, manifest, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating cloud config")
	}

	return nil
}

func (d DirectorImpl) DiffCloudConfig(manifest []byte) (ConfigDiff, error) {
	resp, err := d.client.DiffCloudConfig(manifest)
	if err != nil {
		return ConfigDiff{}, err
	}

	return NewConfigDiff(resp.Diff), nil
}

func (c Client) DiffCloudConfig(manifest []byte) (ConfigDiffResponse, error) {
	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	return c.postConfigDiff("/cloud_configs/diff", manifest, setHeaders)
}
