package manifest

import (
	"net"
	"regexp"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	binet "github.com/cloudfoundry/bosh-cli/common/net"
	boshinst "github.com/cloudfoundry/bosh-cli/installation"
	birelsetmanifest "github.com/cloudfoundry/bosh-cli/release/set/manifest"
)

type Validator interface {
	Validate(Manifest, birelsetmanifest.Manifest) error
	ValidateReleaseJobs(Manifest, boshinst.ReleaseManager) error
}

type validator struct {
	logger boshlog.Logger
}

func NewValidator(logger boshlog.Logger) Validator {
	return &validator{
		logger: logger,
	}
}

func (v *validator) Validate(deploymentManifest Manifest, releaseSetManifest birelsetmanifest.Manifest) error {
	errs := []error{}
	if v.isBlank(deploymentManifest.Name) {
		errs = append(errs, bosherr.Error("name must be provided"))
	}

	networksErrors := v.validateNetworks(deploymentManifest.Networks)
	errs = append(errs, networksErrors...)

	for idx, resourcePool := range deploymentManifest.ResourcePools {
		if v.isBlank(resourcePool.Name) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].name must be provided", idx))
		}
		if v.isBlank(resourcePool.Network) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].network must be provided", idx))
		} else if _, ok := v.networkNames(deploymentManifest)[resourcePool.Network]; !ok {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].network must be the name of a network", idx))
		}

		if v.isBlank(resourcePool.Stemcell.URL) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].stemcell.url must be provided", idx))
		}

		matched, err := regexp.MatchString("^(file|http|https)://", resourcePool.Stemcell.URL)
		if err != nil || !matched {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].stemcell.url must be a valid URL (file:// or http(s)://)", idx))
		}

		if strings.HasPrefix(resourcePool.Stemcell.URL, "http") && v.isBlank(resourcePool.Stemcell.SHA1) {
			errs = append(errs, bosherr.Errorf("resource_pools[%d].stemcell.sha1 must be provided for http URL", idx))
		}
	}

	for idx, diskPool := range deploymentManifest.DiskPools {
		if v.isBlank(diskPool.Name) {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].name must be provided", idx))
		}
		if diskPool.DiskSize <= 0 {
			errs = append(errs, bosherr.Errorf("disk_pools[%d].disk_size must be > 0", idx))
		}
	}

	if len(deploymentManifest.Jobs) > 1 {
		errs = append(errs, bosherr.Error("jobs must be of size 1"))
	}

	for idx, job := range deploymentManifest.Jobs {
		if v.isBlank(job.Name) {
			errs = append(errs, bosherr.Errorf("jobs[%d].name must be provided", idx))
		}
		if job.PersistentDisk < 0 {
			errs = append(errs, bosherr.Errorf("jobs[%d].persistent_disk must be >= 0", idx))
		}
		if job.PersistentDiskPool != "" {
			if _, ok := v.diskPoolNames(deploymentManifest)[job.PersistentDiskPool]; !ok {
				errs = append(errs, bosherr.Errorf("jobs[%d].persistent_disk_pool must be the name of a disk pool", idx))
			}
		}
		if job.Instances < 0 {
			errs = append(errs, bosherr.Errorf("jobs[%d].instances must be >= 0", idx))
		}
		if len(job.Networks) == 0 {
			errs = append(errs, bosherr.Errorf("jobs[%d].networks must be a non-empty array", idx))
		}
		if v.isBlank(job.ResourcePool) {
			errs = append(errs, bosherr.Errorf("jobs[%d].resource_pool must be provided", idx))
		} else {
			if _, ok := v.resourcePoolNames(deploymentManifest)[job.ResourcePool]; !ok {
				errs = append(errs, bosherr.Errorf("jobs[%d].resource_pool must be the name of a resource pool", idx))
			}
		}

		errs = append(errs, v.validateJobNetworks(job.Networks, deploymentManifest.Networks, idx)...)

		if job.Lifecycle != "" && job.Lifecycle != JobLifecycleService {
			errs = append(errs, bosherr.Errorf("jobs[%d].lifecycle must be 'service' ('%s' not supported)", idx, job.Lifecycle))
		}

		templateNames := map[string]struct{}{}
		for templateIdx, template := range job.Templates {
			if v.isBlank(template.Name) {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].name must be provided", idx, templateIdx))
			}
			if _, found := templateNames[template.Name]; found {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].name '%s' must be unique", idx, templateIdx, template.Name))
			}
			templateNames[template.Name] = struct{}{}

			if v.isBlank(template.Release) {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].release must be provided", idx, templateIdx))
			} else {
				_, found := releaseSetManifest.FindByName(template.Release)
				if !found {
					errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].release '%s' must refer to release in releases", idx, templateIdx, template.Release))
				}
			}
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) ValidateReleaseJobs(deploymentManifest Manifest, releaseManager boshinst.ReleaseManager) error {
	errs := []error{}

	for idx, job := range deploymentManifest.Jobs {
		for templateIdx, template := range job.Templates {
			release, found := releaseManager.Find(template.Release)
			if !found {
				errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d].release '%s' must refer to release in releases", idx, templateIdx, template.Release))
			} else {
				_, found := release.FindJobByName(template.Name)
				if !found {
					errs = append(errs, bosherr.Errorf("jobs[%d].templates[%d] must refer to a job in '%s', but there is no job named '%s'", idx, templateIdx, release.Name(), template.Name))
				}
			}
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}

func (v *validator) isBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func (v *validator) networkNames(deploymentManifest Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, network := range deploymentManifest.Networks {
		names[network.Name] = struct{}{}
	}
	return names
}

func (v *validator) diskPoolNames(deploymentManifest Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, diskPool := range deploymentManifest.DiskPools {
		names[diskPool.Name] = struct{}{}
	}
	return names
}

func (v *validator) resourcePoolNames(deploymentManifest Manifest) map[string]struct{} {
	names := make(map[string]struct{})
	for _, resourcePool := range deploymentManifest.ResourcePools {
		names[resourcePool.Name] = struct{}{}
	}
	return names
}

func (v *validator) isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

type maybeIPNet interface {
	Try(func(*net.IPNet) error) error
}

type nothingIpNet struct{}

func (in *nothingIpNet) Try(fn func(*net.IPNet) error) error {
	return nil
}

type somethingIpNet struct {
	ipNet *net.IPNet
}

func (in *somethingIpNet) Try(fn func(*net.IPNet) error) error {
	return fn(in.ipNet)
}

func (v *validator) validateRange(idx int, ipRange string) ([]error, maybeIPNet) {
	if v.isBlank(ipRange) {
		return []error{bosherr.Errorf("networks[%d].subnets[0].range must be provided", idx)}, &nothingIpNet{}
	}

	_, ipNet, err := net.ParseCIDR(ipRange)
	if err != nil {
		return []error{bosherr.Errorf("networks[%d].subnets[0].range must be an ip range", idx)}, &nothingIpNet{}
	}

	return []error{}, &somethingIpNet{ipNet: ipNet}
}

func (v *validator) validateNetworks(networks []Network) []error {
	errs := []error{}

	for idx, network := range networks {
		networkErrors := v.validateNetwork(network, idx)
		errs = append(errs, networkErrors...)
	}

	return errs
}

func (v *validator) validateNetwork(network Network, networkIdx int) []error {
	errs := []error{}

	if v.isBlank(network.Name) {
		errs = append(errs, bosherr.Errorf("networks[%d].name must be provided", networkIdx))
	}

	if network.Type != Dynamic && network.Type != Manual && network.Type != VIP {
		errs = append(errs, bosherr.Errorf("networks[%d].type must be 'manual', 'dynamic', or 'vip'", networkIdx))
	}

	if network.Type == Manual {
		if len(network.Subnets) != 1 {
			errs = append(errs, bosherr.Errorf("networks[%d].subnets must be of size 1", networkIdx))
		} else {
			ipRange := network.Subnets[0].Range
			rangeErrors, maybeIpNet := v.validateRange(networkIdx, ipRange)
			errs = append(errs, rangeErrors...)

			gateway := network.Subnets[0].Gateway
			gatewayErrors := v.validateGateway(networkIdx, gateway, maybeIpNet)
			errs = append(errs, gatewayErrors...)
		}
	}

	return errs
}

func (v *validator) validateJobNetworks(jobNetworks []JobNetwork, networks []Network, jobIdx int) []error {
	errs := []error{}
	defaultCounts := make(map[NetworkDefault]int)

	for networkIdx, jobNetwork := range jobNetworks {
		if v.isBlank(jobNetwork.Name) {
			errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d].name must be provided", jobIdx, networkIdx))
		}

		var matchingNetwork Network

		found := false

		for _, network := range networks {
			if network.Name == jobNetwork.Name {
				found = true
				matchingNetwork = network
			}
		}

		if !found {
			errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d] not found in networks", jobIdx, networkIdx))
		}

		for ipIdx, ip := range jobNetwork.StaticIPs {
			staticIPErrors := v.validateStaticIP(ip, jobNetwork, matchingNetwork, jobIdx, networkIdx, ipIdx)
			errs = append(errs, staticIPErrors...)
		}

		for defaultIdx, value := range jobNetwork.Defaults {
			if value != NetworkDefaultDNS && value != NetworkDefaultGateway {
				errs = append(errs, bosherr.Errorf("jobs[%d].networks[%d].default[%d] must be 'dns' or 'gateway'", jobIdx, networkIdx, defaultIdx))
			}
		}

		for _, dflt := range jobNetwork.Defaults {
			count, present := defaultCounts[dflt]
			if present {
				defaultCounts[dflt] = count + 1
			} else {
				defaultCounts[dflt] = 1
			}
		}
	}

	for _, dflt := range []NetworkDefault{"dns", "gateway"} {
		count, found := defaultCounts[dflt]
		if len(jobNetworks) > 1 && !found {
			errs = append(errs, bosherr.Errorf("with multiple networks, a default for '%s' must be specified", dflt))
		} else if count > 1 {
			errs = append(errs, bosherr.Errorf("only one network can be the default for '%s'", dflt))
		}
	}

	return errs
}

func (v *validator) validateStaticIP(ip string, jobNetwork JobNetwork, network Network, jobIdx, networkIdx, ipIdx int) []error {
	if !v.isValidIP(ip) {
		return []error{bosherr.Errorf("jobs[%d].networks[%d].static_ips[%d] must be a valid IP", jobIdx, networkIdx, ipIdx)}
	}

	if network.Type != Manual {
		return []error{}
	}

	foundInSubnetRange := false
	for _, subnet := range network.Subnets {
		_, rangeNet, err := net.ParseCIDR(subnet.Range)
		if err == nil && rangeNet.Contains(net.ParseIP(ip)) {
			foundInSubnetRange = true
		}
	}

	if foundInSubnetRange {
		return []error{}
	}

	return []error{bosherr.Errorf("jobs[%d].networks[%d] static ip '%s' must be within subnet range", jobIdx, networkIdx, ip)}
}

func (v *validator) validateGateway(idx int, gateway string, ipNet maybeIPNet) []error {
	if v.isBlank(gateway) {
		return []error{bosherr.Errorf("networks[%d].subnets[0].gateway must be provided", idx)}
	}

	errors := []error{}

	_ = ipNet.Try(func(ipNet *net.IPNet) error {
		gatewayIp := net.ParseIP(gateway)
		if gatewayIp == nil {
			errors = append(errors, bosherr.Errorf("networks[%d].subnets[0].gateway must be an ip", idx))
		}

		if !ipNet.Contains(gatewayIp) {
			errors = append(errors, bosherr.Errorf("subnet gateway '%s' must be within the specified range '%s'", gateway, ipNet))
		}

		if ipNet.IP.Equal(gatewayIp) {
			errors = append(errors, bosherr.Errorf("subnet gateway can't be the network address '%s'", gatewayIp))
		}

		if binet.LastAddress(ipNet).Equal(gatewayIp) {
			errors = append(errors, bosherr.Errorf("subnet gateway can't be the broadcast address '%s'", gatewayIp))
		}

		return nil
	})

	return errors
}
