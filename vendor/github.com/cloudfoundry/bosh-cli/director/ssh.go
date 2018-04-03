package director

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type SSHResult struct {
	Hosts []Host

	GatewayUsername string
	GatewayHost     string
}

type Host struct {
	Job       string
	IndexOrID string

	Username      string
	Host          string
	HostPublicKey string
}

type SSHResp struct {
	Status string

	Job   string
	Index *int
	ID    string

	IP            string // e.g. "10.244.2.18"
	HostPublicKey string `json:"host_public_key"`

	GatewayUser string `json:"gateway_user"`
	GatewayHost string `json:"gateway_host"`
}

func (r SSHResp) IndexOrID() string {
	if len(r.ID) > 0 {
		return r.ID
	}

	if r.Index != nil {
		return strconv.Itoa(*r.Index)
	}

	return ""
}

func (d DeploymentImpl) SetUpSSH(slug AllOrInstanceGroupOrInstanceSlug, opts SSHOpts) (SSHResult, error) {
	var result SSHResult

	resps, err := d.client.SetUpSSH(d.name, slug.Name(), slug.IndexOrID(), opts)
	if err != nil {
		return result, err
	}

	if len(resps) == 0 {
		return result, bosherr.Errorf("Did not create any SSH sessions for the instances '%#v'", resps)
	}

	for _, resp := range resps {
		if resp.Status != "success" {
			// todo how to best clean up all ssh sessions?
			return result, bosherr.Errorf("Failed to set up SSH session for one of the instances '%#v'", resp)
		}

		result.Hosts = append(result.Hosts, Host{
			Job:       resp.Job,
			IndexOrID: resp.IndexOrID(),

			Username:      opts.Username,
			Host:          resp.IP,
			HostPublicKey: resp.HostPublicKey,
		})
	}

	// Assumes that all gw are same
	result.GatewayUsername = resps[0].GatewayUser
	result.GatewayHost = resps[0].GatewayHost

	return result, nil
}

func (d DeploymentImpl) CleanUpSSH(slug AllOrInstanceGroupOrInstanceSlug, opts SSHOpts) error {
	return d.client.CleanUpSSH(d.name, slug.Name(), slug.IndexOrID(), opts)
}

func (c Client) SetUpSSH(deploymentName, jobName, indexOrID string, opts SSHOpts) ([]SSHResp, error) {
	var resps []SSHResp

	if len(deploymentName) == 0 {
		return resps, bosherr.Error("Expected non-empty deployment name")
	}

	// jobName and indexOrID may be empty

	path := fmt.Sprintf("/deployments/%s/ssh", deploymentName)

	body := c.buildSSHBody(deploymentName, jobName, indexOrID)

	body["command"] = "setup"
	body["params"] = map[string]string{
		"user":       opts.Username,
		"public_key": opts.PublicKey,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	resultBytes, err := c.taskClientRequest.PostResult(path, reqBody, setHeaders)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Setting up SSH in deployment '%s'", deploymentName)
	}

	err = json.Unmarshal(resultBytes, &resps)
	if err != nil {
		return resps, bosherr.WrapErrorf(err, "Unmarshaling SSH result")
	}

	return resps, nil
}

func (c Client) CleanUpSSH(deploymentName, jobName, indexOrID string, opts SSHOpts) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	// jobName and indexOrID may be empty

	path := fmt.Sprintf("/deployments/%s/ssh", deploymentName)

	body := c.buildSSHBody(deploymentName, jobName, indexOrID)

	body["command"] = "cleanup"
	body["params"] = map[string]string{"user_regex": "^" + opts.Username}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	// Using clientRequest and not taskClientRequest
	// since we don't want to wait for cleanup to finish.
	_, _, err = c.clientRequest.RawPost(path, reqBody, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Cleaning up SSH in deployment '%s'", deploymentName)
	}

	return nil
}

func (c Client) buildSSHBody(deploymentName, jobName, indexOrID string) map[string]interface{} {
	target := map[string]interface{}{}

	if len(jobName) > 0 {
		target["job"] = jobName
	}

	if len(indexOrID) > 0 {
		target["indexes"] = []string{indexOrID}
		target["ids"] = []string{indexOrID}
	} else {
		target["indexes"] = []string{}
		target["ids"] = []string{}
	}

	return map[string]interface{}{
		"deployment_name": deploymentName,
		"target":          target,
	}
}
