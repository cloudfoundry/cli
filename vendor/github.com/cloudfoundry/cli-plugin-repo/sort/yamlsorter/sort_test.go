package yamlsorter_test

import (
	. "github.com/cloudfoundry/cli-plugin-repo/sort/yamlsorter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorting", func() {
	var sorter YAMLSorter
	var sortedPlugins []byte
	var unsortedPlugins []byte

	BeforeEach(func() {
		sortedPlugins = []byte(`plugins:
- authors:
  - contact: someone@gmail.com
    homepage: google.com
    name: someone
  binaries:
  - checksum: 68e8e06ad853488b681e9c538a889596fcf2b3a2
    platform: osx
    url: https://github.com/odlp/antifreeze/releases/download/v0.3.0/antifreeze-darwin
  - checksum: 2b58d45d936e5a4a6eb6e134913c3c335478cdc4
    platform: win64
    url: https://github.com/odlp/antifreeze/releases/download/v0.3.0/antifreeze.exe
  - checksum: dd63f82673592cd15a048606f4f7bd9ee11254f4
    platform: linux64
    url: https://github.com/odlp/antifreeze/releases/download/v0.3.0/antifreeze-linux
  company: null
  created: 2016-04-23T00:00:00Z
  description: Detect if an app has unexpected ENV vars or services bound which are
    missing from the manifest
  homepage: google.com
  name: antifreeze
  updated: 2016-05-20T00:00:00Z
  version: 0.3.0
- authors:
  - name: bob@gmail.com
  binaries:
  - checksum: b250c6c7e01f9e495d1788728a7d65e21bcda203
    platform: osx
    url: https://github.com/concourse/autopilot/releases/download/0.0.1/autopilot-darwin
  - checksum: 8a5cc1ffcaaf71f3b1e0f83c9de68a8100bc94b1
    platform: win64
    url: https://github.com/concourse/autopilot/releases/download/0.0.1/autopilot.exe
  - checksum: 3aaa4c0a7dc8ffeb277f5a111679709a1a5756c5
    platform: linux64
    url: https://github.com/concourse/autopilot/releases/download/0.0.1/autopilot-linux
  company: Concourse
  created: 2015-03-01T00:00:00Z
  description: |
    1. zero downtime deploy plugin for cf applications
    2. other cool stuff
  homepage: yahoo.com
  name: autopilot
  updated: 2015-03-01T00:00:00Z
  version: 0.0.1
`)

		unsortedPlugins = []byte(`plugins:
- authors:
  - contact: null
    name: bob@gmail.com
  created: 2015-03-01T00:00:00Z
  binaries:
  - checksum: b250c6c7e01f9e495d1788728a7d65e21bcda203
    url: https://github.com/concourse/autopilot/releases/download/0.0.1/autopilot-darwin
    platform: osx
  - url: https://github.com/concourse/autopilot/releases/download/0.0.1/autopilot.exe
    platform: win64
    checksum: 8a5cc1ffcaaf71f3b1e0f83c9de68a8100bc94b1
  - checksum: 3aaa4c0a7dc8ffeb277f5a111679709a1a5756c5
    url: https://github.com/concourse/autopilot/releases/download/0.0.1/autopilot-linux
    platform: linux64
  company: Concourse
  name: autopilot
  version: 0.0.1
  description: |
    1. zero downtime deploy plugin for cf applications
    2. other cool stuff
  homepage: yahoo.com
  updated: 2015-03-01T00:00:00Z
- created: 2016-04-23T00:00:00Z
  authors:
  - contact: someone@gmail.com
    name: someone
    homepage: google.com
  binaries:
  - checksum: 68e8e06ad853488b681e9c538a889596fcf2b3a2
    platform: osx
    url: https://github.com/odlp/antifreeze/releases/download/v0.3.0/antifreeze-darwin
  - checksum: 2b58d45d936e5a4a6eb6e134913c3c335478cdc4
    platform: win64
    url: https://github.com/odlp/antifreeze/releases/download/v0.3.0/antifreeze.exe
  - checksum: dd63f82673592cd15a048606f4f7bd9ee11254f4
    platform: linux64
    url: https://github.com/odlp/antifreeze/releases/download/v0.3.0/antifreeze-linux
  description: Detect if an app has unexpected ENV vars or services bound which are
    missing from the manifest
  homepage: google.com
  name: antifreeze
  updated: 2016-05-20T00:00:00Z
  version: 0.3.0
`)
	})

	Context("", func() {
		It("sorts the plugins", func() {
			newPlugins, err := sorter.Sort(unsortedPlugins)
			Expect(err).ToNot(HaveOccurred())
			Expect(newPlugins).To(Equal(sortedPlugins))
		})
	})
})
