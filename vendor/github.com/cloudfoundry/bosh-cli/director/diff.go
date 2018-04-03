package director

import (
	"fmt"
	"net/http"
	gourl "net/url"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentDiffResponse struct {
	Context map[string]interface{} `json:"context"`
	Diff    [][]interface{}        `json:"diff"`
}

type DiffLines [][]interface{}

type DeploymentDiff struct {
	context map[string]interface{}
	Diff    [][]interface{}
}

func NewDeploymentDiff(diff [][]interface{}, context map[string]interface{}) DeploymentDiff {
	return DeploymentDiff{
		context: context,
		Diff:    diff,
	}
}

func (d DeploymentImpl) Diff(manifest []byte, doNotRedact bool) (DeploymentDiff, error) {
	resp, err := d.client.Diff(manifest, d.name, doNotRedact)
	if err != nil {
		return DeploymentDiff{}, err
	}

	return NewDeploymentDiff(resp.Diff, resp.Context), nil
}

func (c Client) Diff(manifest []byte, deploymentName string, doNotRedact bool) (DeploymentDiffResponse, error) {
	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	query := gourl.Values{}

	if doNotRedact {
		query.Add("redact", "false")
	} else {
		query.Add("redact", "true")
	}

	path := fmt.Sprintf("/deployments/%s/diff?%s", deploymentName, query.Encode())

	var resp DeploymentDiffResponse

	err := c.clientRequest.Post(path, manifest, setHeaders, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Fetching diff result")
	}

	return resp, nil
}
