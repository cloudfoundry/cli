package configv3_test

import (
	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginRepository", func() {
	Describe("PluginRepositories", func() {
		It("returns sorted plugin repositories", func() {
			config := Config{
				ConfigFile: CFConfig{
					PluginRepositories: []PluginRepository{
						{Name: "S-repo", URL: "S-repo.com"},
						{Name: "repo-2", URL: "repo2.com"},
						{Name: "repo-1", URL: "repo1.com"},
					},
				},
			}

			Expect(config.PluginRepositories()).To(Equal([]PluginRepository{
				{Name: "repo-1", URL: "repo1.com"},
				{Name: "repo-2", URL: "repo2.com"},
				{Name: "S-repo", URL: "S-repo.com"},
			}))
		})
	})

	Describe("AddPluginRepository", func() {
		It("adds the repo name and url to the list of repositories", func() {
			config := Config{
				ConfigFile: CFConfig{
					PluginRepositories: []PluginRepository{
						{Name: "S-repo", URL: "S-repo.com"},
						{Name: "repo-2", URL: "repo2.com"},
						{Name: "repo-1", URL: "repo1.com"},
					},
				},
			}

			config.AddPluginRepository("some-repo", "some-URL")
			Expect(config.PluginRepositories()).To(ContainElement(PluginRepository{Name: "some-repo", URL: "some-URL"}))
		})
	})
})
