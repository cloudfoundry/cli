package configuration_test

import (
	. "cf/configuration"
	"fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	"os"
	"path/filepath"
)

func withFakeHome(t mr.TestingT, callback func(dirPath string)) {
	fileutils.TempDir("test-config", func(dir string, err error) {
		if err != nil {
			Fail("Couldn't create tmp file")
		}
		callback(filepath.Join(dir, ".cf", "config.json"))
	})
}

func withConfigFixture(t mr.TestingT, name string, callback func(dirPath string)) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	callback(filepath.Join(cwd, "../../fixtures/config", name, ".cf", "config.json"))
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestLoadingWithNoConfigFile", func() {
		withFakeHome(mr.T(), func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(configData.Target).To(Equal(""))
			Expect(configData.ApiVersion).To(Equal(""))
			Expect(configData.AuthorizationEndpoint).To(Equal(""))
			Expect(configData.AccessToken).To(Equal(""))
		})
	})

	It("TestSavingAndLoading", func() {
		withFakeHome(mr.T(), func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()
			Expect(err).NotTo(HaveOccurred())

			configData.ApiVersion = "3.1.0"
			configData.Target = "https://api.target.example.com"
			configData.AuthorizationEndpoint = "https://login.target.example.com"
			configData.AccessToken = "bearer my_access_token"

			err = repo.Save(configData)
			Expect(err).NotTo(HaveOccurred())

			savedConfig, err := repo.Load()
			Expect(err).NotTo(HaveOccurred())
			Expect(savedConfig).To(Equal(configData))
		})
	})

	It("TestReadingOutdatedConfigReturnsNewConfig", func() {
		withConfigFixture(mr.T(), "outdated-config", func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(configData.Target).To(Equal(""))
		})
	})
})
