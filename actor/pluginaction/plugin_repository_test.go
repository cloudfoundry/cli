package pluginaction_test

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plugin Repository Actions", func() {
	var (
		actor            *Actor
		fakeConfig       *pluginactionfakes.FakeConfig
		fakePluginClient *pluginactionfakes.FakePluginClient
	)

	BeforeEach(func() {
		fakeConfig = new(pluginactionfakes.FakeConfig)
		fakePluginClient = new(pluginactionfakes.FakePluginClient)
		actor = NewActor(fakeConfig, fakePluginClient)
	})

	Describe("AddPluginRepository", func() {
		var err error

		JustBeforeEach(func() {
			err = actor.AddPluginRepository("some-repo", "some-URL")
		})

		Context("when passed a url without a scheme", func() {
			It("prepends https://", func() {
				_ = actor.AddPluginRepository("some-repo2", "some-URL")
				url := fakePluginClient.GetPluginRepositoryArgsForCall(1)
				Expect(strings.HasPrefix(url, "https://")).To(BeTrue())
			})
		})

		Context("when passed a schemeless IP address with a port", func() {
			It("prepends https://", func() {
				_ = actor.AddPluginRepository("some-repo2", "127.0.0.1:5000")
				url := fakePluginClient.GetPluginRepositoryArgsForCall(1)
				Expect(strings.HasPrefix(url, "https://")).To(BeTrue())
			})
		})

		Context("when the repository name is taken", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{
						Name: "repo-1",
						URL:  "https://URL-1",
					},
					{
						Name: "some-repo",
						URL:  "https://www.com",
					},
				})
			})

			It("returns the RepositoryNameTakenError", func() {
				Expect(err).To(MatchError(actionerror.RepositoryNameTakenError{Name: "some-repo"}))
			})
		})

		Context("when the repository name and URL are already taken in the same repo", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{
						Name: "some-repo",
						URL:  "https://some-URL",
					},
				})
			})

			It("returns a RepositoryAlreadyExistsError", func() {
				Expect(err).To(MatchError(actionerror.RepositoryAlreadyExistsError{Name: "some-repo", URL: "https://some-URL"}))

				Expect(fakePluginClient.GetPluginRepositoryCallCount()).To(Equal(0))
				Expect(fakeConfig.AddPluginRepositoryCallCount()).To(Equal(0))
			})
		})

		Context("when the repository name is the same except for case sensitivity", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{
						Name: "sOmE-rEpO",
						URL:  "https://some-URL",
					},
				})
			})

			It("returns a RepositoryAlreadyExistsError", func() {
				Expect(err).To(MatchError(actionerror.RepositoryAlreadyExistsError{Name: "sOmE-rEpO", URL: "https://some-URL"}))

				Expect(fakePluginClient.GetPluginRepositoryCallCount()).To(Equal(0))
				Expect(fakeConfig.AddPluginRepositoryCallCount()).To(Equal(0))
			})
		})

		Context("when the repository name is the same and repostiroy URL is the same except for trailing slash", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{
						Name: "some-repo",
						URL:  "https://some-URL",
					},
				})
			})

			It("returns a RepositoryAlreadyExistsError", func() {
				err = actor.AddPluginRepository("some-repo", "some-URL/")
				Expect(err).To(MatchError(actionerror.RepositoryAlreadyExistsError{Name: "some-repo", URL: "https://some-URL"}))

				Expect(fakePluginClient.GetPluginRepositoryCallCount()).To(Equal(0))
				Expect(fakeConfig.AddPluginRepositoryCallCount()).To(Equal(0))
			})
		})

		Context("when the repository URL is taken", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{
						Name: "repo-1",
						URL:  "https://URL-1",
					},
					{
						Name: "repo-2",
						URL:  "https://some-URL",
					},
				})
			})

			It("adds the repo to the config and returns nil", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginClient.GetPluginRepositoryCallCount()).To(Equal(1))
				Expect(fakePluginClient.GetPluginRepositoryArgsForCall(0)).To(Equal("https://some-URL"))

				Expect(fakeConfig.AddPluginRepositoryCallCount()).To(Equal(1))
				repoName, repoURL := fakeConfig.AddPluginRepositoryArgsForCall(0)
				Expect(repoName).To(Equal("some-repo"))
				Expect(repoURL).To(Equal("https://some-URL"))
			})
		})

		Context("when getting the repository errors", func() {
			BeforeEach(func() {
				fakePluginClient.GetPluginRepositoryReturns(plugin.PluginRepository{}, errors.New("generic-error"))
			})

			It("returns an AddPluginRepositoryError", func() {
				Expect(err).To(MatchError(actionerror.AddPluginRepositoryError{
					Name:    "some-repo",
					URL:     "https://some-URL",
					Message: "generic-error",
				}))
			})
		})

		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakePluginClient.GetPluginRepositoryReturns(plugin.PluginRepository{}, nil)
			})

			It("adds the repo to the config and returns nil", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginClient.GetPluginRepositoryCallCount()).To(Equal(1))
				Expect(fakePluginClient.GetPluginRepositoryArgsForCall(0)).To(Equal("https://some-URL"))

				Expect(fakeConfig.AddPluginRepositoryCallCount()).To(Equal(1))
				repoName, repoURL := fakeConfig.AddPluginRepositoryArgsForCall(0)
				Expect(repoName).To(Equal("some-repo"))
				Expect(repoURL).To(Equal("https://some-URL"))
			})
		})
	})

	Describe("GetPluginRepository", func() {
		Context("when the repository is registered", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{Name: "some-REPO", URL: "some-url"},
				})
			})

			It("returns the repository case-insensitively", func() {
				pluginRepo, err := actor.GetPluginRepository("sOmE-rEpO")
				Expect(err).ToNot(HaveOccurred())
				Expect(pluginRepo).To(Equal(configv3.PluginRepository{Name: "some-REPO", URL: "some-url"}))
			})
		})

		Context("when the repository is not registered", func() {
			It("returns a RepositoryNotRegisteredError", func() {
				_, err := actor.GetPluginRepository("some-rEPO")
				Expect(err).To(MatchError(actionerror.RepositoryNotRegisteredError{Name: "some-rEPO"}))
			})
		})
	})

	Describe("IsPluginRepositoryRegistered", func() {
		Context("when the repository is registered", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{
					{Name: "some-repo"},
				})
			})

			It("returns true", func() {
				Expect(actor.IsPluginRepositoryRegistered("some-repo")).To(BeTrue())
			})
		})

		Context("when the repository is not registered", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{})
			})

			It("returns true", func() {
				Expect(actor.IsPluginRepositoryRegistered("some-repo")).To(BeFalse())
			})
		})
	})
})
