package manifest

import (
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type ResourcePool struct {
	Name            string
	Network         string
	CloudProperties biproperty.Map
	Env             biproperty.Map
	Stemcell        StemcellRef
}

type StemcellRef struct {
	URL  string
	SHA1 string
}

func (s StemcellRef) GetURL() string {
	return s.URL
}

func (s StemcellRef) GetSHA1() string {
	return s.SHA1
}

func (s StemcellRef) Description() string {
	return "stemcell"
}
