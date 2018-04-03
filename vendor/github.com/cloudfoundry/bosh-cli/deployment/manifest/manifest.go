package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Manifest struct {
	Name          string
	Properties    biproperty.Map
	Jobs          []Job
	Networks      []Network
	DiskPools     []DiskPool
	ResourcePools []ResourcePool
	Update        Update
	Tags          map[string]string
}

type Update struct {
	UpdateWatchTime WatchTime
}

// NetworkInterfaces returns a map of network names to network interfaces.
// We can't use map[string]NetworkInterface, because it's impossible to down-cast to what the cloud client requires.
//TODO: refactor to NetworkInterfaces(Job) and use FindJobByName before using (then remove error)
func (d Manifest) NetworkInterfaces(jobName string) (map[string]biproperty.Map, error) {
	job, found := d.FindJobByName(jobName)
	if !found {
		return map[string]biproperty.Map{}, bosherr.Errorf("Could not find job with name: %s", jobName)
	}

	networkMap := d.networkMap()

	ifaceMap := map[string]biproperty.Map{}
	var err error
	for _, jobNetwork := range job.Networks {
		network := networkMap[jobNetwork.Name]
		ifaceMap[jobNetwork.Name], err = network.Interface(jobNetwork.StaticIPs, jobNetwork.Defaults)
		if err != nil {
			return map[string]biproperty.Map{}, bosherr.WrapError(err, "Building network interface")
		}
	}
	if len(job.Networks) == 1 {
		ifaceMap[job.Networks[0].Name]["default"] = []NetworkDefault{NetworkDefaultDNS, NetworkDefaultGateway}
	}

	return ifaceMap, nil
}

func (d Manifest) JobName() string {
	// Currently we deploy only one job
	return d.Jobs[0].Name
}

func (d Manifest) Stemcell(jobName string) (StemcellRef, error) {
	resourcePool, err := d.ResourcePool(jobName)
	if err != nil {
		return StemcellRef{}, err
	}
	return resourcePool.Stemcell, nil
}

func (d Manifest) ResourcePool(jobName string) (ResourcePool, error) {
	job, found := d.FindJobByName(jobName)
	if !found {
		return ResourcePool{}, bosherr.Errorf("Could not find job with name: %s", jobName)
	}

	for _, resourcePool := range d.ResourcePools {
		if resourcePool.Name == job.ResourcePool {
			return resourcePool, nil
		}
	}
	err := bosherr.Errorf("Could not find resource pool '%s' for job '%s'", job.ResourcePool, jobName)
	return ResourcePool{}, err
}

func (d Manifest) DiskPool(jobName string) (DiskPool, error) {
	job, found := d.FindJobByName(jobName)
	if !found {
		return DiskPool{}, bosherr.Errorf("Could not find job with name: %s", jobName)
	}

	if job.PersistentDiskPool != "" {
		for _, diskPool := range d.DiskPools {
			if diskPool.Name == job.PersistentDiskPool {
				return diskPool, nil
			}
		}
		err := bosherr.Errorf("Could not find persistent disk pool '%s' for job '%s'", job.PersistentDiskPool, jobName)
		return DiskPool{}, err
	}

	if job.PersistentDisk > 0 {
		diskPool := DiskPool{
			DiskSize:        job.PersistentDisk,
			CloudProperties: biproperty.Map{},
		}
		return diskPool, nil
	}

	return DiskPool{}, nil
}

func (d Manifest) networkMap() map[string]Network {
	result := map[string]Network{}
	for _, network := range d.Networks {
		result[network.Name] = network
	}
	return result
}

func (d Manifest) FindJobByName(jobName string) (Job, bool) {
	for _, job := range d.Jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return Job{}, false
}

func (d Manifest) GetListOfTemplateReleases() (map[string]string, bool) {
	if len(d.Jobs) != 1 {
		return nil, false
	} else {
		result := make(map[string]string)

		for _, job := range d.Jobs[0].Templates {
			result[job.Release] = job.Release
		}

		return result, true
	}
}
