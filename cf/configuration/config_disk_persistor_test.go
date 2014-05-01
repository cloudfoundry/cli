/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package configuration_test

import (
	. "github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/gofileutils/fileutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

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

var _ = Describe("Testing with ginkgo", func() {
	It("TestLoadingWithNoConfigFile", func() {
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

	It("TestSavingAndLoading", func() {
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

	It("TestReadingOutdatedConfigReturnsNewConfig", func() {
		withConfigFixture("outdated-config", func(configPath string) {
			repo := NewDiskPersistor(configPath)
			configData, err := repo.Load()

			Expect(err).NotTo(HaveOccurred())
			Expect(configData.Target).To(Equal(""))
		})
	})
})
