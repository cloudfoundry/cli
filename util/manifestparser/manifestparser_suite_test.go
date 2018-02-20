package manifestparser_test

import (
	"io/ioutil"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

func TestManifestparser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manifest Parser Suite")
}

func WriteManifest(path string, manifest map[string]interface{}) {
	body, err := yaml.Marshal(manifest)
	Expect(err).ToNot(HaveOccurred())
	err = ioutil.WriteFile(path, body, 0666)
	Expect(err).ToNot(HaveOccurred())
}
