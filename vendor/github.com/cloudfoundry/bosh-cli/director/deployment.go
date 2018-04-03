package director

import (
	"encoding/json"
	"fmt"
	"net/http"
	gourl "net/url"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DeploymentImpl struct {
	client Client

	name        string
	cloudConfig string

	manifest string

	releases  []Release
	stemcells []Stemcell
	teams     []string

	fetched  bool
	fetchErr error
}

type ExportReleaseResult struct {
	BlobstoreID string
	SHA1        string
}

type ExportReleaseResp struct {
	BlobstoreID string `json:"blobstore_id"`
	SHA1        string `json:"sha1"`
}

type LogsResult struct {
	BlobstoreID string
	SHA1        string
}

type VariableResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (d DeploymentImpl) Name() string { return d.name }

func (d *DeploymentImpl) CloudConfig() (string, error) {
	d.fetch()
	return d.cloudConfig, d.fetchErr
}

func (d *DeploymentImpl) Releases() ([]Release, error) {
	d.fetch()
	return d.releases, d.fetchErr
}

func (d *DeploymentImpl) Stemcells() ([]Stemcell, error) {
	d.fetch()
	return d.stemcells, d.fetchErr
}

func (d *DeploymentImpl) Teams() ([]string, error) {
	d.fetch()
	return d.teams, d.fetchErr
}

func (d *DeploymentImpl) fetch() {
	if d.fetched {
		return
	}

	resps, err := d.client.Deployments()
	if err != nil {
		d.fetchErr = err
		return
	}

	for _, resp := range resps {
		if resp.Name == d.name {
			d.fill(resp)
			return
		}
	}

	d.fetchErr = bosherr.Errorf("Expected to find deployment '%s'", d.name)
}

func (d *DeploymentImpl) fill(resp DeploymentResp) {
	d.fetched = true

	rels, err := newReleasesFromResps(resp.Releases, d.client)
	if err != nil {
		d.fetchErr = err
		return
	}

	stems, err := newStemcellsFromResps(resp.Stemcells, d.client)
	if err != nil {
		d.fetchErr = err
		return
	}

	d.releases = rels
	d.stemcells = stems
	d.teams = resp.Teams
	d.cloudConfig = resp.CloudConfig
}

func (d DeploymentImpl) Manifest() (string, error) {
	resp, err := d.client.Deployment(d.name)
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Fetching manifest")
	}

	return resp.Manifest, nil
}

func (d DeploymentImpl) FetchLogs(slug AllOrInstanceGroupOrInstanceSlug, filters []string, agent bool) (LogsResult, error) {
	blobID, sha1, err := d.client.FetchLogs(d.name, slug.Name(), slug.IndexOrID(), filters, agent)
	if err != nil {
		return LogsResult{}, err
	}

	return LogsResult{BlobstoreID: blobID, SHA1: sha1}, nil
}

func (d DeploymentImpl) EnableResurrection(slug InstanceSlug, enabled bool) error {
	return d.client.EnableResurrection(d.name, slug.Name(), slug.IndexOrID(), enabled)
}

func (d DeploymentImpl) Ignore(slug InstanceSlug, enabled bool) error {
	return d.client.Ignore(d.name, slug.Name(), slug.IndexOrID(), enabled)
}

func (d DeploymentImpl) Start(slug AllOrInstanceGroupOrInstanceSlug, opts StartOpts) error {
	return d.changeJobState("started", slug, false, false, false, false, opts.Canaries, opts.MaxInFlight)
}

func (d DeploymentImpl) Stop(slug AllOrInstanceGroupOrInstanceSlug, opts StopOpts) error {
	if opts.Hard {
		return d.changeJobState("detached", slug, opts.SkipDrain, opts.Force, false, false, opts.Canaries, opts.MaxInFlight)
	}
	return d.changeJobState("stopped", slug, opts.SkipDrain, opts.Force, false, false, opts.Canaries, opts.MaxInFlight)
}

func (d DeploymentImpl) Restart(slug AllOrInstanceGroupOrInstanceSlug, opts RestartOpts) error {
	return d.changeJobState("restart", slug, opts.SkipDrain, opts.Force, false, false, opts.Canaries, opts.MaxInFlight)
}

func (d DeploymentImpl) Recreate(slug AllOrInstanceGroupOrInstanceSlug, opts RecreateOpts) error {
	return d.changeJobState("recreate", slug, opts.SkipDrain, opts.Force, opts.Fix, opts.DryRun, opts.Canaries, opts.MaxInFlight)
}

func (d DeploymentImpl) changeJobState(state string, slug AllOrInstanceGroupOrInstanceSlug, skipDrain bool, force bool, fix bool, dryRun bool, canaries string, maxInFlight string) error {
	return d.client.ChangeJobState(
		state, d.name, slug.Name(), slug.IndexOrID(), skipDrain, force, fix, dryRun, canaries, maxInFlight)
}

func (d DeploymentImpl) ExportRelease(release ReleaseSlug, os OSVersionSlug, jobs []string) (ExportReleaseResult, error) {
	resp, err := d.client.ExportRelease(d.name, release, os, jobs)
	if err != nil {
		return ExportReleaseResult{}, err
	}

	return ExportReleaseResult{BlobstoreID: resp.BlobstoreID, SHA1: resp.SHA1}, nil
}

func (d DeploymentImpl) Update(manifest []byte, opts UpdateOpts) error {
	return d.client.UpdateDeployment(manifest, opts)
}

func (d DeploymentImpl) Delete(force bool) error {
	err := d.client.DeleteDeployment(d.name, force)
	if err != nil {
		resps, listErr := d.client.Deployments()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.Name == d.name {
				return err
			}
		}
	}

	return nil
}

func (d DeploymentImpl) AttachDisk(slug InstanceSlug, diskCID string) error {
	values := gourl.Values{}
	values.Add("deployment", d.Name())
	values.Add("job", slug.Name())
	values.Add("instance_id", slug.IndexOrID())

	path := fmt.Sprintf("/disks/%s/attachments?%s", diskCID, values.Encode())
	_, err := d.client.taskClientRequest.PutResult(path, []byte{}, func(*http.Request) {})
	return err
}

func (d DeploymentImpl) IsInProgress() (bool, error) {
	lockResps, err := d.client.Locks()
	if err != nil {
		return false, err
	}

	for _, r := range lockResps {
		if r.IsForDeployment(d.name) {
			return true, nil
		}
	}

	return false, nil
}

func (d DeploymentImpl) Variables() ([]VariableResult, error) {
	path := fmt.Sprintf("/deployments/%s/variables", d.name)
	response := []VariableResult{}

	if err := d.client.clientRequest.Get(path, &response); err != nil {
		return nil, bosherr.WrapErrorf(err, "Error fetching variables for deployment '%s'", d.name)
	}

	return response, nil
}

func (c Client) FetchLogs(deploymentName, job, indexOrID string, filters []string, agent bool) (string, string, error) {
	if len(deploymentName) == 0 {
		return "", "", bosherr.Error("Expected non-empty deployment name")
	}

	if len(job) == 0 {
		job = "*"
	}

	if len(indexOrID) == 0 {
		indexOrID = "*"
	}

	query := gourl.Values{}

	if len(filters) > 0 {
		query.Add("filters", strings.Join(filters, ","))
	}

	if agent {
		query.Add("type", "agent")
	} else {
		query.Add("type", "job")
	}

	path := fmt.Sprintf("/deployments/%s/jobs/%s/%s/logs?%s",
		deploymentName, job, indexOrID, query.Encode())

	taskID, _, err := c.taskClientRequest.GetResult(path)
	if err != nil {
		return "", "", bosherr.WrapErrorf(err, "Fetching logs")
	}

	taskResp, err := c.Task(taskID)
	if err != nil {
		return "", "", err
	}

	return taskResp.Result, "", nil
}

func (c Client) Ignore(deploymentName, instanceGroup, indexOrID string, enabled bool) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	if len(instanceGroup) == 0 {
		return bosherr.Error("Expected non-empty instance group name")
	}

	if len(indexOrID) == 0 {
		return bosherr.Error("Expected non-empty index or ID")
	}

	headers := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	body := map[string]bool{"ignore": enabled}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	path := fmt.Sprintf("/deployments/%s/instance_groups/%s/%s/ignore",
		deploymentName, instanceGroup, indexOrID)

	_, _, err = c.clientRequest.RawPut(path, reqBody, headers)
	if err != nil {
		msg := "Changing ignore state for '%s/%s' in deployment '%s'"
		return bosherr.WrapErrorf(err, msg, instanceGroup, indexOrID, deploymentName)
	}

	return nil
}

func (c Client) EnableResurrection(deploymentName, job, indexOrID string, enabled bool) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	if len(job) == 0 {
		return bosherr.Error("Expected non-empty job name")
	}

	if len(indexOrID) == 0 {
		return bosherr.Error("Expected non-empty index or ID")
	}

	path := fmt.Sprintf("/deployments/%s/jobs/%s/%s/resurrection",
		deploymentName, job, indexOrID)

	body := map[string]bool{"resurrection_paused": !enabled}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, _, err = c.clientRequest.RawPut(path, reqBody, setHeaders)
	if err != nil {
		msg := "Changing VM resurrection state for '%s/%s' in deployment '%s'"
		return bosherr.WrapErrorf(err, msg, job, indexOrID, deploymentName)
	}

	return nil
}

func (c Client) ChangeJobState(state, deploymentName, job, indexOrID string, skipDrain bool, force bool, fix bool, dryRun bool, canaries string, maxInFlight string) error {
	if len(state) == 0 {
		return bosherr.Error("Expected non-empty job state")
	}

	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	// allows to have empty job and indexOrID

	query := gourl.Values{}

	query.Add("state", state)

	if skipDrain {
		query.Add("skip_drain", "true")
	}

	if force {
		query.Add("force", "true")
	}

	if fix {
		query.Add("fix", "true")
	}

	if dryRun {
		query.Add("dry_run", "true")
	}

	if canaries != "" {
		query.Add("canaries", canaries)
	}

	if maxInFlight != "" {
		query.Add("max_in_flight", maxInFlight)
	}

	path := fmt.Sprintf("/deployments/%s/jobs", deploymentName)

	if len(job) > 0 {
		path += "/" + job

		if len(indexOrID) > 0 {
			path += "/" + indexOrID
		}
	} else {
		path += "/*"
	}

	path += "?" + query.Encode()

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	_, err := c.taskClientRequest.PutResult(path, []byte{}, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Changing state")
	}

	return nil
}

func (c Client) ExportRelease(deploymentName string, release ReleaseSlug, os OSVersionSlug, jobs []string) (ExportReleaseResp, error) {
	var resp ExportReleaseResp

	if len(deploymentName) == 0 {
		return resp, bosherr.Error("Expected non-empty deployment name")
	}

	if len(release.Name()) == 0 {
		return resp, bosherr.Error("Expected non-empty release name")
	}

	if len(release.Version()) == 0 {
		return resp, bosherr.Error("Expected non-empty release version")
	}

	if len(os.OS()) == 0 {
		return resp, bosherr.Error("Expected non-empty OS name")
	}

	if len(os.Version()) == 0 {
		return resp, bosherr.Error("Expected non-empty OS version")
	}

	jobFilters := []map[string]string{}
	for _, job := range jobs {
		jobFilters = append(jobFilters, map[string]string{"name": job})
	}

	path := "/releases/export"

	body := map[string]interface{}{
		"deployment_name":  deploymentName,
		"release_name":     release.Name(),
		"release_version":  release.Version(),
		"stemcell_os":      os.OS(),
		"stemcell_version": os.Version(),
		"sha2":             true,
		"jobs":             jobFilters,
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Marshaling request body")
	}

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	resultBytes, err := c.taskClientRequest.PostResult(path, reqBody, setHeaders)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Exporting release")
	}

	err = json.Unmarshal(resultBytes, &resp)
	if err != nil {
		return resp, bosherr.WrapErrorf(err, "Unmarshaling export release result")
	}

	return resp, nil
}

func (c Client) UpdateDeployment(manifest []byte, opts UpdateOpts) error {
	query := gourl.Values{}

	if opts.Recreate {
		query.Add("recreate", "true")
	}

	if opts.Fix {
		query.Add("fix", "true")
	}

	if len(opts.SkipDrain.AsQueryValue()) > 0 {
		query.Add("skip_drain", opts.SkipDrain.AsQueryValue())
	}

	if opts.Canaries != "" {
		query.Add("canaries", opts.Canaries)
	}

	if opts.MaxInFlight != "" {
		query.Add("max_in_flight", opts.MaxInFlight)
	}

	if opts.DryRun {
		query.Add("dry_run", "true")
	}

	if len(opts.Diff.context) != 0 {
		context := map[string]interface{}{}

		for key, value := range opts.Diff.context {
			context[key] = value
		}

		contextJson, err := json.Marshal(context)
		if err != nil {
			return bosherr.WrapErrorf(err, "Marshaling context")
		}

		query.Add("context", string(contextJson))
	}

	path := fmt.Sprintf("/deployments?%s", query.Encode())

	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "text/yaml")
	}

	_, err := c.taskClientRequest.PostResult(path, manifest, setHeaders)
	if err != nil {
		return bosherr.WrapErrorf(err, "Updating deployment")
	}

	return nil
}

func (c Client) DeleteDeployment(deploymentName string, force bool) error {
	if len(deploymentName) == 0 {
		return bosherr.Error("Expected non-empty deployment name")
	}

	query := gourl.Values{}

	if force {
		query.Add("force", "true")
	}

	path := fmt.Sprintf("/deployments/%s?%s", deploymentName, query.Encode())

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting deployment '%s'", deploymentName)
	}

	return nil
}

type DeploymentVMResp struct {
	JobName  string `json:"job"`   // e.g. dummy1
	JobIndex int    `json:"index"` // e.g. 0,1,2

	AgentID string `json:"agent_id"` // e.g. 3b30123e-dfa6-4eff-abe6-63c2d5a88938
	CID     string // e.g. vm-ce10ae6a-6c31-413b-a134-7179f49e0bda
}

func (c Client) DeploymentVMs(deploymentName string) ([]DeploymentVMResp, error) {
	if len(deploymentName) == 0 {
		return nil, bosherr.Error("Expected non-empty deployment name")
	}

	var vms []DeploymentVMResp

	path := fmt.Sprintf("/deployments/%s/vms", deploymentName)

	err := c.clientRequest.Get(path, &vms)
	if err != nil {
		return vms, bosherr.WrapErrorf(err, "Listing deployment '%s' VMs", deploymentName)
	}

	return vms, nil
}
