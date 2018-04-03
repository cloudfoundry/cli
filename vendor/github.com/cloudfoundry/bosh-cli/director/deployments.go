package director

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	semver "github.com/cppforlife/go-semi-semantic/version"
)

type DeploymentResp struct {
	Name string

	Manifest string

	Releases  []DeploymentReleaseResp
	Stemcells []DeploymentStemcellResp
	Teams     []string

	CloudConfig string `json:"cloud_config"`
}

type DeploymentReleaseResp struct {
	Name    string
	Version string
}

type DeploymentStemcellResp struct {
	Name    string
	Version string
}

func (d DirectorImpl) Deployments() ([]Deployment, error) {
	deps := []Deployment{}

	resps, err := d.client.Deployments()
	if err != nil {
		return deps, err
	}

	for _, resp := range resps {
		dep := &DeploymentImpl{client: d.client, name: resp.Name}

		dep.fill(resp)

		deps = append(deps, dep)
	}

	return deps, nil
}

func (d DirectorImpl) FindDeployment(name string) (Deployment, error) {
	if len(name) == 0 {
		return nil, bosherr.Error("Expected non-empty deployment name")
	}

	return &DeploymentImpl{client: d.client, name: name}, nil
}

func (c Client) Deployments() ([]DeploymentResp, error) {
	var deps []DeploymentResp

	err := c.clientRequest.Get("/deployments", &deps)
	if err != nil {
		return deps, bosherr.WrapErrorf(err, "Finding deployments")
	}

	return deps, nil
}

func (c Client) Deployment(name string) (DeploymentResp, error) {
	var resp DeploymentResp

	if len(name) == 0 {
		return resp, bosherr.Error("Expected non-empty deployment name")
	}

	path := fmt.Sprintf("/deployments/%s", name)

	err := c.clientRequest.Get(path, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Finding deployment '%s'", name)
	}

	return resp, nil
}

func newReleasesFromResps(resps []DeploymentReleaseResp, client Client) ([]Release, error) {
	var rels []Release

	for _, resp := range resps {
		parsedVersion, err := semver.NewVersionFromString(resp.Version)
		if err != nil {
			return nil, bosherr.WrapErrorf(
				err, "Parsing version for release '%s/%s'", resp.Name, resp.Version)
		}

		rel := &ReleaseImpl{
			client:  client,
			name:    resp.Name,
			version: parsedVersion,
		}

		rels = append(rels, rel)
	}

	return rels, nil
}

func newStemcellsFromResps(resps []DeploymentStemcellResp, client Client) ([]Stemcell, error) {
	var stems []Stemcell

	for _, resp := range resps {
		parsedVersion, err := semver.NewVersionFromString(resp.Version)
		if err != nil {
			return nil, bosherr.WrapErrorf(
				err, "Parsing version for stemcell '%s/%s'", resp.Name, resp.Version)
		}

		stemcell := StemcellImpl{
			client:  client,
			name:    resp.Name,
			version: parsedVersion,
		}

		stems = append(stems, stemcell)
	}

	return stems, nil
}
