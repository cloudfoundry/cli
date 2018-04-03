package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type DiskPool struct {
	Name            string
	DiskSize        int
	CloudProperties biproperty.Map
}
