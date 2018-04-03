package installation

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	boshrel "github.com/cloudfoundry/bosh-cli/release"
)

type ReleaseManager interface {
	Add(boshrel.Release)
	List() []boshrel.Release
	Find(string) (boshrel.Release, bool)
	DeleteAll() error
}

type releaseManager struct {
	releases []boshrel.Release

	logTag string
	logger boshlog.Logger
}

func NewReleaseManager(logger boshlog.Logger) ReleaseManager {
	return &releaseManager{
		releases: []boshrel.Release{},

		logTag: "installation.releaseManager",
		logger: logger,
	}
}

func (m *releaseManager) Add(release boshrel.Release) {
	m.logger.Info(m.logTag, "Adding extracted release '%s-%s'", release.Name(), release.Version())
	m.releases = append(m.releases, release)
}

func (m *releaseManager) List() []boshrel.Release {
	return append([]boshrel.Release(nil), m.releases...)
}

func (m *releaseManager) Find(name string) (boshrel.Release, bool) {
	for _, release := range m.releases {
		if release.Name() == name {
			return release, true
		}
	}

	return nil, false
}

func (m *releaseManager) DeleteAll() error {
	for _, release := range m.releases {
		deleteErr := release.CleanUp()
		if deleteErr != nil {
			return bosherr.Errorf("Failed to delete extracted release '%s': %s", release.Name(), deleteErr.Error())
		}
	}

	m.releases = []boshrel.Release{}

	return nil
}
