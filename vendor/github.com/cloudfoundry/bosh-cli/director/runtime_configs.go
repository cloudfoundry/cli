package director

import (
	"fmt"
	"net/http"

	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type RuntimeConfig struct {
	Properties string
}

func (d DirectorImpl) LatestRuntimeConfig(name string) (RuntimeConfig, error) {
	resps, err := d.client.RuntimeConfigs(name)
	if err != nil {
		return RuntimeConfig{}, err
	}

	if len(resps) == 0 {
		return RuntimeConfig{}, bosherr.Error("No runtime config")
	}

	return resps[0], nil
}

func (d DirectorImpl) UpdateRuntimeConfig(name string, manifest []byte) error {
	return d.client.UpdateRuntimeConfig(name, manifest)
}

func (c Client) RuntimeConfigs(name string) ([]RuntimeConfig, error) {
	var resps []RuntimeConfig

	query := gourl.Values{}
	query.Add("name", name)
	query.Add("limit", "1")

	path := fmt.Sprintf("/runtime_configs?%s", query.Encode())

	err := c.clientRequest.Get(path, &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Finding runtime configs")
	}

	return resps, nil
}

func (c Client) UpdateRuntimeConfig(name string, manifest []byte) error {
	query := gourl.Values{}
	query.Add("name", name)

	path := fmt.Sprintf("/runtime_configs?%s", query.Encode())

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	_, _, err := c.clientRequest.RawPost(path, manifest, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating runtime config")
	}

	return nil
}

func (d DirectorImpl) DiffRuntimeConfig(name string, manifest []byte, noRedact bool) (ConfigDiff, error) {
	resp, err := d.client.DiffRuntimeConfig(name, manifest, noRedact)
	if err != nil {
		return ConfigDiff{}, err
	}

	return NewConfigDiff(resp.Diff), nil
}

func (c Client) DiffRuntimeConfig(name string, manifest []byte, noRedact bool) (ConfigDiffResponse, error) {
	query := gourl.Values{}
	query.Add("name", name)

	if noRedact {
		query.Add("redact", "false")
	}

	path := fmt.Sprintf("/runtime_configs/diff?%s", query.Encode())

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	return c.postConfigDiff(path, manifest, setHeaders)
}
