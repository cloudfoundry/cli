package manifest

import (
	boshman "github.com/cloudfoundry/bosh-cli/release/manifest"
)

type Manifest struct {
	Releases []boshman.ReleaseRef
}

func (d Manifest) ReleasesByName() map[string]boshman.ReleaseRef {
	releasesByName := map[string]boshman.ReleaseRef{}
	for _, release := range d.Releases {
		releasesByName[release.Name] = release
	}
	return releasesByName
}

func (d Manifest) FindByName(name string) (boshman.ReleaseRef, bool) {
	for _, release := range d.Releases {
		if release.Name == name {
			return release, true
		}
	}
	return boshman.ReleaseRef{}, false
}
