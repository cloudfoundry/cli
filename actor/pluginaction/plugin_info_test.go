package pluginaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("plugin info actions", func() {
	var (
		actor      *Actor
		fakeClient *pluginactionfakes.FakePluginClient
	)

	BeforeEach(func() {
		fakeClient = new(pluginactionfakes.FakePluginClient)
		actor = NewActor(nil, fakeClient)
	})

	Describe("GetPluginInfoFromRepositoriesForPlatform", func() {
		Context("when there is a single repository", func() {
			Context("when getting the plugin repository errors", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryReturns(plugin.PluginRepository{}, errors.New("some-error"))
				})

				It("returns a FetchingPluginInfoFromRepositoryError", func() {
					_, _, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", []configv3.PluginRepository{{Name: "some-repository", URL: "some-url"}}, "some-platform")
					Expect(err).To(MatchError(actionerror.FetchingPluginInfoFromRepositoryError{
						RepositoryName: "some-repository",
						Err:            errors.New("some-error"),
					}))
				})
			})

			Context("when getting the plugin repository succeeds", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryReturns(plugin.PluginRepository{
						Plugins: []plugin.Plugin{
							{
								Name:    "some-plugin",
								Version: "1.2.3",
								Binaries: []plugin.PluginBinary{
									{Platform: "osx", URL: "http://some-darwin-url", Checksum: "somechecksum"},
									{Platform: "win64", URL: "http://some-windows-url", Checksum: "anotherchecksum"},
									{Platform: "linux64", URL: "http://some-linux-url", Checksum: "lastchecksum"},
								},
							},
							{
								Name:    "linux-plugin",
								Version: "1.5.0",
								Binaries: []plugin.PluginBinary{
									{Platform: "osx", URL: "http://some-url", Checksum: "somechecksum"},
									{Platform: "win64", URL: "http://another-url", Checksum: "anotherchecksum"},
									{Platform: "linux64", URL: "http://last-url", Checksum: "lastchecksum"},
								},
							},
							{
								Name:    "osx-plugin",
								Version: "3.0.0",
								Binaries: []plugin.PluginBinary{
									{Platform: "osx", URL: "http://some-url", Checksum: "somechecksum"},
									{Platform: "win64", URL: "http://another-url", Checksum: "anotherchecksum"},
									{Platform: "linux64", URL: "http://last-url", Checksum: "lastchecksum"},
								},
							},
						},
					}, nil)
				})

				Context("when the specified plugin does not exist in the repository", func() {
					It("returns a PluginNotFoundInRepositoryError", func() {
						_, _, err := actor.GetPluginInfoFromRepositoriesForPlatform("plugin-i-dont-exist", []configv3.PluginRepository{{Name: "some-repo", URL: "some-url"}}, "platform-i-dont-exist")
						Expect(err).To(MatchError(actionerror.PluginNotFoundInAnyRepositoryError{
							PluginName: "plugin-i-dont-exist",
						}))
					})
				})

				Context("when the specified plugin for the provided platform does not exist in the repository", func() {
					It("returns a NoCompatibleBinaryError", func() {
						_, _, err := actor.GetPluginInfoFromRepositoriesForPlatform("linux-plugin", []configv3.PluginRepository{{Name: "some-repo", URL: "some-url"}}, "platform-i-dont-exist")
						Expect(err).To(MatchError(actionerror.NoCompatibleBinaryError{}))
					})
				})

				Context("when the specified plugin exists", func() {
					It("returns the plugin info", func() {
						pluginInfo, repos, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", []configv3.PluginRepository{{Name: "some-repo", URL: "some-url"}}, "osx")
						Expect(err).ToNot(HaveOccurred())
						Expect(pluginInfo.Name).To(Equal("some-plugin"))
						Expect(pluginInfo.Version).To(Equal("1.2.3"))
						Expect(pluginInfo.URL).To(Equal("http://some-darwin-url"))
						Expect(repos).To(ConsistOf("some-repo"))
					})
				})
			})
		})

		Context("when there are multiple repositories", func() {
			var pluginRepositories []configv3.PluginRepository

			BeforeEach(func() {
				pluginRepositories = []configv3.PluginRepository{
					{Name: "repo1", URL: "url1"},
					{Name: "repo2", URL: "url2"},
					{Name: "repo3", URL: "url3"},
				}
			})

			Context("when getting a plugin repository errors", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryReturnsOnCall(0, plugin.PluginRepository{
						Plugins: []plugin.Plugin{
							{
								Name:    "some-plugin",
								Version: "1.2.3",
								Binaries: []plugin.PluginBinary{
									{Platform: "osx", URL: "http://some-darwin-url", Checksum: "somechecksum"},
									{Platform: "win64", URL: "http://some-windows-url", Checksum: "anotherchecksum"},
									{Platform: "linux64", URL: "http://some-linux-url", Checksum: "lastchecksum"},
								},
							},
						},
					}, nil)
					fakeClient.GetPluginRepositoryReturnsOnCall(1, plugin.PluginRepository{}, errors.New("some-error"))
				})

				It("returns a FetchingPluginInfoFromRepositoryError", func() {
					_, _, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")
					Expect(err).To(MatchError(actionerror.FetchingPluginInfoFromRepositoryError{
						RepositoryName: "repo2",
						Err:            errors.New("some-error")}))
				})
			})

			Context("when the plugin isn't found", func() {
				It("returns the PluginNotFoundInAnyRepositoryError", func() {
					_, _, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")
					Expect(err).To(Equal(actionerror.PluginNotFoundInAnyRepositoryError{PluginName: "some-plugin"}))
				})
			})

			Context("when no compatible binaries are found for the plugin", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryStub = func(repoURL string) (plugin.PluginRepository, error) {
						return plugin.PluginRepository{Plugins: []plugin.Plugin{
							{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
								{Platform: "incompatible-platform", URL: "some-url", Checksum: "some-checksum"},
							}},
						}}, nil
					}
				})

				It("returns the NoCompatibleBinaryError", func() {
					_, _, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")
					Expect(err).To(MatchError(actionerror.NoCompatibleBinaryError{}))
				})
			})

			Context("when some binaries are compatible and some are not", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryStub = func(repoURL string) (plugin.PluginRepository, error) {
						if repoURL == "url1" {
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "incompatible-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						} else {
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						}
					}
				})

				It("returns the compatible plugin info and a list of the repositories it was found in", func() {
					pluginInfo, repos, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")

					Expect(err).ToNot(HaveOccurred())
					Expect(pluginInfo).To(Equal(PluginInfo{
						Name:     "some-plugin",
						Version:  "1.2.3",
						URL:      "some-url",
						Checksum: "some-checksum",
					}))
					Expect(repos).To(ConsistOf("repo2", "repo3"))
				})
			})

			Context("when the plugin is found in one repository", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryStub = func(repoURL string) (plugin.PluginRepository, error) {
						if repoURL == "url1" {
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						} else {
							return plugin.PluginRepository{}, nil
						}
					}
				})

				It("returns the plugin info and a list of the repositories it was found in", func() {
					pluginInfo, repos, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")

					Expect(err).ToNot(HaveOccurred())
					Expect(pluginInfo).To(Equal(PluginInfo{
						Name:     "some-plugin",
						Version:  "1.2.3",
						URL:      "some-url",
						Checksum: "some-checksum",
					}))
					Expect(repos).To(ConsistOf("repo1"))
				})
			})

			Context("when the plugin is found in many repositories", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryStub = func(repoURL string) (plugin.PluginRepository, error) {
						return plugin.PluginRepository{Plugins: []plugin.Plugin{
							{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
								{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
							}},
						}}, nil
					}
				})

				It("returns the plugin info and a list of the repositories it was found in", func() {
					pluginInfo, repos, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")

					Expect(err).ToNot(HaveOccurred())
					Expect(pluginInfo).To(Equal(PluginInfo{
						Name:     "some-plugin",
						Version:  "1.2.3",
						URL:      "some-url",
						Checksum: "some-checksum",
					}))
					Expect(repos).To(ConsistOf("repo1", "repo2", "repo3"))
				})
			})

			Context("when different versions of the plugin are found in all the repositories", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryStub = func(repoURL string) (plugin.PluginRepository, error) {
						switch repoURL {
						case "url1":
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						case "url2":
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "2.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						default:
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "0.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						}
					}
				})

				It("returns the newest plugin info and only the repository it was found in", func() {
					pluginInfo, repos, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")

					Expect(err).ToNot(HaveOccurred())
					Expect(pluginInfo).To(Equal(PluginInfo{
						Name:     "some-plugin",
						Version:  "2.2.3",
						URL:      "some-url",
						Checksum: "some-checksum",
					}))
					Expect(repos).To(ConsistOf("repo2"))
				})
			})

			Context("when some repositories contain a newer version of the plugin than others", func() {
				BeforeEach(func() {
					fakeClient.GetPluginRepositoryStub = func(repoURL string) (plugin.PluginRepository, error) {
						switch repoURL {
						case "url1", "url2":
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "1.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						default:
							return plugin.PluginRepository{Plugins: []plugin.Plugin{
								{Name: "some-plugin", Version: "0.2.3", Binaries: []plugin.PluginBinary{
									{Platform: "some-platform", URL: "some-url", Checksum: "some-checksum"},
								}},
							}}, nil
						}
					}
				})

				It("returns only the newest plugin info and the list of repositories it's contained in", func() {
					pluginInfo, repos, err := actor.GetPluginInfoFromRepositoriesForPlatform("some-plugin", pluginRepositories, "some-platform")

					Expect(err).ToNot(HaveOccurred())
					Expect(pluginInfo).To(Equal(PluginInfo{
						Name:     "some-plugin",
						Version:  "1.2.3",
						URL:      "some-url",
						Checksum: "some-checksum",
					}))
					Expect(repos).To(ConsistOf("repo1", "repo2"))
				})
			})
		})
	})
})
