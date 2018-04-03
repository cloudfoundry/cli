package manifest

import (
	biutil "github.com/cloudfoundry/bosh-cli/common/util"
	bidepltpl "github.com/cloudfoundry/bosh-cli/deployment/template"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"
)

type Parser interface {
	Parse(interpolatedTemplate bidepltpl.InterpolatedTemplate, path string) (Manifest, error)
}

type parser struct {
	fs     boshsys.FileSystem
	logger boshlog.Logger
	logTag string
}

type manifest struct {
	Name           string
	Update         UpdateSpec
	Networks       []network
	ResourcePools  []resourcePool `yaml:"resource_pools"`
	DiskPools      []diskPool     `yaml:"disk_pools"`
	Jobs           []job
	InstanceGroups []job `yaml:"instance_groups"`
	Properties     map[interface{}]interface{}
	Tags           map[string]string
}

type UpdateSpec struct {
	UpdateWatchTime *string `yaml:"update_watch_time"`
}

type network struct {
	Name            string                      `yaml:"name"`
	Type            string                      `yaml:"type"`
	CloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
	Subnets         []subnet                    `yaml:"subnets"`
	IP              string                      `yaml:"ip"`
	Netmask         string                      `yaml:"netmask"`
	Gateway         string                      `yaml:"gateway"`
	DNS             []string                    `yaml:"dns"`
}

type subnet struct {
	Range           string                      `yaml:"range"`
	Gateway         string                      `yaml:"gateway"`
	DNS             []string                    `yaml:"dns"`
	CloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

type resourcePool struct {
	Name            string                      `yaml:"name"`
	Network         string                      `yaml:"network"`
	CloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
	Env             map[interface{}]interface{} `yaml:"env"`
	Stemcell        stemcellRef                 `yaml:"stemcell"`
}

type diskPool struct {
	Name            string                      `yaml:"name"`
	DiskSize        int                         `yaml:"disk_size"`
	CloudProperties map[interface{}]interface{} `yaml:"cloud_properties"`
}

type job struct {
	Name               string
	Instances          int
	Lifecycle          string
	Templates          []releaseJobRef
	Jobs               []releaseJobRef `yaml:"jobs"`
	Networks           []jobNetwork
	PersistentDisk     int    `yaml:"persistent_disk"`
	PersistentDiskPool string `yaml:"persistent_disk_pool"`
	ResourcePool       string `yaml:"resource_pool"`
	Properties         map[interface{}]interface{}
}

type releaseJobRef struct {
	Name    string
	Release string

	// This is a pointer so we can differentiate between `properties: {}`
	// and not specifying the key at all.
	Properties *map[interface{}]interface{}
}

type stemcellRef struct {
	URL  string
	SHA1 string
}

type jobNetwork struct {
	Name      string
	Defaults  []string `yaml:"default"`
	StaticIPs []string `yaml:"static_ips"`
}

var boshDeploymentDefaults = Manifest{
	Update: Update{
		UpdateWatchTime: WatchTime{
			Start: 0,
			End:   300000,
		},
	},
}

func NewParser(fs boshsys.FileSystem, logger boshlog.Logger) Parser {
	return &parser{
		fs:     fs,
		logger: logger,
		logTag: "deploymentParser",
	}
}

func (p *parser) Parse(interpolatedTemplate bidepltpl.InterpolatedTemplate, path string) (Manifest, error) {
	bytes := interpolatedTemplate.Content()

	comboManifest := manifest{}

	err := yaml.Unmarshal(bytes, &comboManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}

	p.logger.Debug(p.logTag, "Parsed BOSH deployment manifest: %#v", comboManifest)

	deploymentManifest, err := p.parseDeploymentManifest(comboManifest, path)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling BOSH deployment manifest")
	}

	return deploymentManifest, nil
}

func (p *parser) parseDeploymentManifest(depManifest manifest, path string) (Manifest, error) {
	deployment := boshDeploymentDefaults
	deployment.Name = depManifest.Name
	deployment.Tags = depManifest.Tags

	networks, err := p.parseNetworkManifests(depManifest.Networks)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing networks: %#v", depManifest.Networks)
	}
	deployment.Networks = networks

	resourcePools, err := p.parseResourcePoolManifests(depManifest.ResourcePools, path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing resource_pools: %#v", depManifest.ResourcePools)
	}

	deployment.ResourcePools = resourcePools

	diskPools, err := p.parseDiskPoolManifests(depManifest.DiskPools)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing disk_pools: %#v", depManifest.DiskPools)
	}
	deployment.DiskPools = diskPools

	if len(depManifest.Jobs) > 0 && len(depManifest.InstanceGroups) > 0 {
		return Manifest{}, bosherr.Error("Deployment specifies both jobs and instance_groups keys, only one is allowed")
	}

	rawJobs := depManifest.Jobs
	if len(depManifest.InstanceGroups) > 0 {
		rawJobs = depManifest.InstanceGroups
	}
	jobs, err := p.parseJobManifests(rawJobs)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing jobs: %#v", depManifest.Jobs)
	}
	deployment.Jobs = jobs

	properties, err := biproperty.BuildMap(depManifest.Properties)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing global manifest properties: %#v", depManifest.Properties)
	}
	deployment.Properties = properties

	if depManifest.Update.UpdateWatchTime != nil {
		updateWatchTime, err := NewWatchTime(*depManifest.Update.UpdateWatchTime)
		if err != nil {
			return Manifest{}, bosherr.WrapError(err, "Parsing update watch time")
		}

		deployment.Update = Update{
			UpdateWatchTime: updateWatchTime,
		}
	}

	return deployment, nil
}

func (p *parser) parseJobManifests(rawJobs []job) ([]Job, error) {
	jobs := make([]Job, len(rawJobs), len(rawJobs))
	for i, rawJob := range rawJobs {
		job := Job{
			Name:               rawJob.Name,
			Instances:          rawJob.Instances,
			Lifecycle:          JobLifecycle(rawJob.Lifecycle),
			PersistentDisk:     rawJob.PersistentDisk,
			PersistentDiskPool: rawJob.PersistentDiskPool,
			ResourcePool:       rawJob.ResourcePool,
		}

		if len(rawJob.Templates) > 0 && len(rawJob.Jobs) > 0 {
			return jobs, bosherr.Error("Deployment specifies both templates and jobs keys for instance_group " + job.Name + ", only one is allowed")
		}

		templates := rawJob.Templates
		if len(rawJob.Jobs) > 0 {
			templates = rawJob.Jobs
		}

		if templates != nil {
			releaseJobRefs := make([]ReleaseJobRef, len(templates), len(templates))
			for i, rawJobRef := range templates {
				ref := ReleaseJobRef{
					Name:    rawJobRef.Name,
					Release: rawJobRef.Release,
				}

				if rawJobRef.Properties != nil {
					properties, err := biproperty.BuildMap(*rawJobRef.Properties)
					if err != nil {
						return []Job{}, bosherr.WrapErrorf(err, "Parsing release job properties: %#v", rawJobRef.Properties)
					}

					ref.Properties = &properties
				}

				releaseJobRefs[i] = ref
			}
			job.Templates = releaseJobRefs
		}

		if rawJob.Networks != nil {
			jobNetworks := make([]JobNetwork, len(rawJob.Networks), len(rawJob.Networks))
			for i, rawJobNetwork := range rawJob.Networks {
				jobNetwork := JobNetwork{
					Name:      rawJobNetwork.Name,
					StaticIPs: rawJobNetwork.StaticIPs,
				}

				if rawJobNetwork.Defaults != nil {
					networkDefaults := make([]NetworkDefault, len(rawJobNetwork.Defaults), len(rawJobNetwork.Defaults))
					for i, rawDefaults := range rawJobNetwork.Defaults {
						networkDefaults[i] = NetworkDefault(rawDefaults)
					}
					jobNetwork.Defaults = networkDefaults
				}

				jobNetworks[i] = jobNetwork
			}
			job.Networks = jobNetworks
		}

		if rawJob.Properties != nil {
			properties, err := biproperty.BuildMap(rawJob.Properties)
			if err != nil {
				return jobs, bosherr.WrapErrorf(err, "Parsing job '%s' properties: %#v", rawJob.Name, rawJob.Properties)
			}
			job.Properties = properties
		}

		jobs[i] = job
	}

	return jobs, nil
}

func (p *parser) parseNetworkManifests(rawNetworks []network) ([]Network, error) {
	networks := make([]Network, len(rawNetworks), len(rawNetworks))
	for i, rawNetwork := range rawNetworks {
		network := Network{
			Name: rawNetwork.Name,
			Type: NetworkType(rawNetwork.Type),
			DNS:  rawNetwork.DNS,
		}

		cloudProperties, err := biproperty.BuildMap(rawNetwork.CloudProperties)
		if err != nil {
			return networks, bosherr.WrapErrorf(err, "Parsing network '%s' cloud_properties: %#v", rawNetwork.Name, rawNetwork.CloudProperties)
		}
		network.CloudProperties = cloudProperties

		for _, subnet := range rawNetwork.Subnets {
			cloudProperties, err := biproperty.BuildMap(subnet.CloudProperties)
			if err != nil {
				return networks, bosherr.WrapErrorf(err, "Parsing network subnet '%s' cloud_properties: %#v", rawNetwork.Name, subnet.CloudProperties)
			}

			network.Subnets = append(network.Subnets, Subnet{
				Range:           subnet.Range,
				Gateway:         subnet.Gateway,
				DNS:             subnet.DNS,
				CloudProperties: cloudProperties,
			})
		}

		networks[i] = network
	}

	return networks, nil
}

func (p *parser) parseResourcePoolManifests(rawResourcePools []resourcePool, path string) ([]ResourcePool, error) {
	resourcePools := make([]ResourcePool, len(rawResourcePools), len(rawResourcePools))
	for i, rawResourcePool := range rawResourcePools {
		resourcePool := ResourcePool{
			Name:     rawResourcePool.Name,
			Network:  rawResourcePool.Network,
			Stemcell: StemcellRef(rawResourcePool.Stemcell),
		}

		cloudProperties, err := biproperty.BuildMap(rawResourcePool.CloudProperties)
		if err != nil {
			return resourcePools, bosherr.WrapErrorf(err, "Parsing resource_pool '%s' cloud_properties: %#v", rawResourcePool.Name, rawResourcePool.CloudProperties)
		}
		resourcePool.CloudProperties = cloudProperties

		env, err := biproperty.BuildMap(rawResourcePool.Env)
		if err != nil {
			return resourcePools, bosherr.WrapErrorf(err, "Parsing resource_pool '%s' env: %#v", rawResourcePool.Name, rawResourcePool.Env)
		}
		resourcePool.Env = env

		resourcePool.Stemcell.URL, err = biutil.AbsolutifyPath(path, resourcePool.Stemcell.URL, p.fs)
		if err != nil {
			return resourcePools, bosherr.WrapErrorf(err, "Resolving stemcell path '%s", resourcePool.Stemcell.URL)
		}

		resourcePools[i] = resourcePool
	}

	return resourcePools, nil
}

func (p *parser) parseDiskPoolManifests(rawDiskPools []diskPool) ([]DiskPool, error) {
	diskPools := make([]DiskPool, len(rawDiskPools), len(rawDiskPools))
	for i, rawDiskPool := range rawDiskPools {
		diskPool := DiskPool{
			Name:     rawDiskPool.Name,
			DiskSize: rawDiskPool.DiskSize,
		}

		cloudProperties, err := biproperty.BuildMap(rawDiskPool.CloudProperties)
		if err != nil {
			return diskPools, bosherr.WrapErrorf(err, "Parsing disk_pool '%s' cloud_properties: %#v", rawDiskPool.Name, rawDiskPool.CloudProperties)
		}
		diskPool.CloudProperties = cloudProperties

		diskPools[i] = diskPool
	}

	return diskPools, nil
}
