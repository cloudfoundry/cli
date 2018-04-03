package manifest

import (
	"fmt"
	"net"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type NetworkType string

func (n NetworkType) String() string {
	return string(n)
}

const (
	Dynamic NetworkType = "dynamic"
	Manual  NetworkType = "manual"
	VIP     NetworkType = "vip"
)

type Network struct {
	Name            string
	Type            NetworkType
	CloudProperties biproperty.Map
	DNS             []string
	Subnets         []Subnet
}

type Subnet struct {
	Range           string
	Gateway         string
	DNS             []string
	CloudProperties biproperty.Map
}

// Interface returns a property map representing a generic network interface.
// Expected Keys: ip, type, cloud properties.
// Optional Keys: netmask, gateway, dns
func (n Network) Interface(staticIPs []string, networkDefaults []NetworkDefault) (biproperty.Map, error) {
	networkInterface := biproperty.Map{
		"type": n.Type.String(),
	}

	if n.Type == Manual {
		networkInterface["gateway"] = n.Subnets[0].Gateway
		if len(n.Subnets[0].DNS) > 0 {
			networkInterface["dns"] = n.Subnets[0].DNS
		}

		_, ipNet, err := net.ParseCIDR(n.Subnets[0].Range)
		if err != nil {
			return biproperty.Map{}, bosherr.WrapError(err, "Failed to parse subnet range")
		}

		networkInterface["netmask"] = ipMaskString(ipNet.Mask)
		networkInterface["cloud_properties"] = n.Subnets[0].CloudProperties
	} else {
		networkInterface["cloud_properties"] = n.CloudProperties
	}

	if n.Type == Dynamic && len(n.DNS) > 0 {
		networkInterface["dns"] = n.DNS
	}

	if len(staticIPs) > 0 {
		networkInterface["ip"] = staticIPs[0]
	}

	if len(networkDefaults) > 0 {
		networkInterface["default"] = networkDefaults
	}

	return networkInterface, nil
}

func ipMaskString(ipMask net.IPMask) string {
	ip := net.IP(ipMask)

	if p4 := ip.To4(); len(p4) == net.IPv4len {
		return ip.String()
	}

	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
		[]byte(ip[0:2]), []byte(ip[2:4]), []byte(ip[4:6]), []byte(ip[6:8]),
		[]byte(ip[8:10]), []byte(ip[10:12]), []byte(ip[12:14]), []byte(ip[14:16]))
}
