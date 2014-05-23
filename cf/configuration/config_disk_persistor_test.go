package configuration_test

import (
	. "github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = Describe("DiskPersistor", func() {
	It("has sane defaults when there is no config to read", func() {
		withFakeHome(func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(configData.Target).To(Equal(""))
			Expect(configData.ApiVersion).To(Equal(""))
			Expect(configData.AuthorizationEndpoint).To(Equal(""))
			Expect(configData.AccessToken).To(Equal(""))
		})
	})

	It("saves its config to disk and can read it back", func() {
		withFakeHome(func(configPath string) {
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

	Context("when the configuration version is older than the current version", func() {
		It("returns a new empty config", func() {
			withConfigFixture("outdated-config", func(configPath string) {
				repo := NewDiskPersistor(configPath)
				configData, err := repo.Load()

				Expect(err).NotTo(HaveOccurred())
				Expect(configData.Target).To(Equal(""))
			})
		})
	})
})

func withFakeHome(callback func(dirPath string)) {
	fileutils.TempDir("test-config", func(dir string, err error) {
		if err != nil {
			Fail("Couldn't create tmp file")
		}
		callback(filepath.Join(dir, ".cf", "config.json"))
	})
}

func withConfigFixture(name string, callback func(dirPath string)) {
	cwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	callback(filepath.Join(cwd, "../../fixtures/config", name, ".cf", "config.json"))
}
