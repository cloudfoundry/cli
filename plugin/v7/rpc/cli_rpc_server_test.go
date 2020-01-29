// +build V7

package rpc_test

import (
	"errors"
	"net"
	"net/rpc"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	plugin "code.cloudfoundry.org/cli/plugin/v7"
	plugin_models "code.cloudfoundry.org/cli/plugin/v7/models"
	cmdRunner "code.cloudfoundry.org/cli/plugin/v7/rpc"
	"code.cloudfoundry.org/cli/plugin/v7/rpc/rpcfakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Server", func() {

	var (
		err        error
		client     *rpc.Client
		rpcService *cmdRunner.CliRpcService
	)

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
	})

	BeforeEach(func() {
		rpc.DefaultServer = rpc.NewServer()
	})

	Describe(".NewRpcService", func() {
		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, rpc.DefaultServer, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an err of another Rpc process is already registered", func() {
			_, err := cmdRunner.NewRpcService(nil, rpc.DefaultServer, nil, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Stop", func() {
		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, rpc.DefaultServer, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
		})

		It("shuts down the rpc server", func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)

			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".Start", func() {
		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, rpc.DefaultServer, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		It("Start an Rpc server for communication", func() {
			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe(".SetPluginMetadata", func() {
		var (
			metadata *plugin.PluginMetadata
		)

		BeforeEach(func() {
			rpcService, err = cmdRunner.NewRpcService(nil, rpc.DefaultServer, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())

			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())

			metadata = &plugin.PluginMetadata{
				Name: "foo",
				Commands: []plugin.Command{
					{Name: "cmd_1", HelpText: "cm 1 help text"},
					{Name: "cmd_2", HelpText: "cmd 2 help text"},
				},
			}
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		It("set the rpc command's Return Data", func() {
			var success bool
			err = client.Call("CliRpcCmd.SetPluginMetadata", metadata, &success)

			Expect(err).ToNot(HaveOccurred())
			Expect(success).To(BeTrue())
			Expect(rpcService.RpcCmd.PluginMetadata).To(Equal(metadata))
		})
	})

	Describe("Plugin API", func() {
		var (
			fakePluginActor   *rpcfakes.FakePluginActor
			fakeConfig        *commandfakes.FakeConfig
			fakeCommandParser *rpcfakes.FakeCommandParser
		)

		BeforeEach(func() {
			fakePluginActor = new(rpcfakes.FakePluginActor)
			fakeConfig = new(commandfakes.FakeConfig)
			fakeCommandParser = new(rpcfakes.FakeCommandParser)

			rpcService, err = cmdRunner.NewRpcService(nil, rpc.DefaultServer, fakeConfig,
				fakePluginActor, fakeCommandParser)
			Expect(err).ToNot(HaveOccurred())

			err := rpcService.Start()
			Expect(err).ToNot(HaveOccurred())

			pingCli(rpcService.Port())
			client, err = rpc.Dial("tcp", "127.0.0.1:"+rpcService.Port())
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			rpcService.Stop()

			//give time for server to stop
			time.Sleep(50 * time.Millisecond)
		})

		Describe(".ApiEndpoint", func() {
			BeforeEach(func() {
				fakeConfig.TargetReturns("www.example.com")
			})

			It("returns the ApiEndpoint() setting in config", func() {
				var result string
				err = client.Call("CliRpcCmd.ApiEndpoint", "", &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal("www.example.com"))
				Expect(fakeConfig.TargetCallCount()).To(Equal(1))
			})
		})

		Describe("CliCommand", func() {
			BeforeEach(func() {
				fakeCommandParser.ParseCommandFromArgsStub = func(ui *ui.UI, args []string) int {
					ui.DisplayText("some-cf-command output")
					return 0
				}
			})

			It("calls a core command", func() {
				var result []string
				err := client.Call("CliRpcCmd.CliCommand", []string{"some-cf-command"}, &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeCommandParser.ParseCommandFromArgsCallCount()).To(Equal(1))
				_, args := fakeCommandParser.ParseCommandFromArgsArgsForCall(0)
				Expect(args).To(Equal([]string{"some-cf-command"}))
				Expect(result).To(Equal([]string{"some-cf-command output"}))
			})
		})

		Describe("GetApp", func() {
			var (
				summary v7action.DetailedApplicationSummary
			)
			BeforeEach(func() {
				summary = v7action.DetailedApplicationSummary{
					ApplicationSummary: v7action.ApplicationSummary{
						Application: v7action.Application{
							GUID:      "some-app-guid",
							Name:      "some-app",
							StackName: "some-stack",
							State:     constant.ApplicationStarted,
						},
						ProcessSummaries: v7action.ProcessSummaries{
							{
								Process: v7action.Process{
									Type:               constant.ProcessTypeWeb,
									Command:            *types.NewFilteredString("some-command-1"),
									MemoryInMB:         types.NullUint64{IsSet: true, Value: 512},
									DiskInMB:           types.NullUint64{IsSet: true, Value: 64},
									HealthCheckTimeout: 60,
									Instances:          types.NullInt{IsSet: true, Value: 5},
								},
								InstanceDetails: []v7action.ProcessInstance{
									{State: constant.ProcessInstanceRunning},
									{State: constant.ProcessInstanceRunning},
									{State: constant.ProcessInstanceCrashed},
									{State: constant.ProcessInstanceRunning},
									{State: constant.ProcessInstanceRunning},
								},
							},
							{
								Process: v7action.Process{
									Type:               "console",
									Command:            *types.NewFilteredString("some-command-2"),
									MemoryInMB:         types.NullUint64{IsSet: true, Value: 256},
									DiskInMB:           types.NullUint64{IsSet: true, Value: 16},
									HealthCheckTimeout: 120,
									Instances:          types.NullInt{IsSet: true, Value: 1},
								},
								InstanceDetails: []v7action.ProcessInstance{
									{State: constant.ProcessInstanceRunning},
								},
							},
						},
					},
					CurrentDroplet: v7action.Droplet{
						Stack: "cflinuxfs2",
						Buildpacks: []v7action.DropletBuildpack{
							{
								Name:         "ruby_buildpack",
								DetectOutput: "some-detect-output",
							},
							{
								Name:         "some-buildpack",
								DetectOutput: "",
							},
						},
					},
				}
				fakePluginActor.GetDetailedAppSummaryReturns(summary, v7action.Warnings{"warning-1", "warning-2"}, nil)

				fakeConfig.HasTargetedOrganizationReturns(true)

				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "some-space",
					GUID: "some-space-guid",
				})

			})

			It("retrieves the app summary", func() {
				result := plugin_models.DetailedApplicationSummary{}
				err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetDetailedAppSummaryCallCount()).To(Equal(1))
				appName, spaceGUID, withObfuscatedValues := fakePluginActor.GetDetailedAppSummaryArgsForCall(0)
				Expect(appName).To(Equal("some-app"))
				Expect(spaceGUID).To(Equal("some-space-guid"))
				Expect(withObfuscatedValues).To(BeTrue())
			})

			It("populates the plugin model with the retrieved app", func() {
				result := plugin_models.DetailedApplicationSummary{}
				err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
				Expect(err).ToNot(HaveOccurred())

				//fmt.Fprintf(os.Stdout, "%+v", result)
				Expect(result).To(BeEquivalentTo(summary))
			})

			Context("when retrieving the app fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetDetailedAppSummaryReturns(v7action.DetailedApplicationSummary{}, nil, errors.New("some-error"))
				})
				It("returns an error", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when no org is targeted", func() {
				BeforeEach(func() {
					fakeConfig.HasTargetedOrganizationReturns(false)
				})
				It("complains that no org is targeted", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
					Expect(err).To(MatchError("no organization targeted"))
				})
			})

			Context("when no space is targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedSpaceReturns(configv3.Space{
						Name: "",
						GUID: "",
					})
				})
				It("complains that no space is targeted", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
					Expect(err).To(MatchError("no space targeted"))
				})
			})
		})

		Describe("GetApps", func() {
			Context("when a space is targeted", func() {
				BeforeEach(func() {
					appList := []v7action.Application{
						v7action.Application{Name: "name1", GUID: "guid1"},
						v7action.Application{Name: "name2", GUID: "guid2"},
						v7action.Application{Name: "name3", GUID: "guid3"},
					}
					fakePluginActor.GetApplicationsBySpaceReturns(appList, v7action.Warnings{"warning-1", "warning-2"}, nil)

					fakeConfig.HasTargetedOrganizationReturns(true)

					fakeConfig.TargetedSpaceReturns(configv3.Space{
						Name: "some-space",
						GUID: "some-space-guid",
					})

				})

				It("retrieves the app summary", func() {
					result := []plugin_models.Application{}
					err := client.Call("CliRpcCmd.GetApps", "", &result)
					Expect(err).ToNot(HaveOccurred())

					Expect(fakePluginActor.GetApplicationsBySpaceCallCount()).To(Equal(1))
					spaceGUID := fakePluginActor.GetApplicationsBySpaceArgsForCall(0)
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(len(result)).To(Equal(3))
				})
			})

			Context("when no org is targeted", func() {
				BeforeEach(func() {
					fakeConfig.HasTargetedOrganizationReturns(false)
				})
				It("complains that no org is targeted", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
					Expect(err).To(MatchError("no organization targeted"))
				})
			})

			Context("when no space is targeted", func() {
				BeforeEach(func() {
					fakeConfig.HasTargetedOrganizationReturns(true)
					fakeConfig.TargetedSpaceReturns(configv3.Space{
						Name: "",
						GUID: "",
					})
				})
				It("complains that no space is targeted", func() {
					result := plugin_models.DetailedApplicationSummary{}
					err := client.Call("CliRpcCmd.GetApp", "some-app", &result)
					Expect(err).To(MatchError("no space targeted"))
				})
			})
		})

		Describe("GetOrg", func() {
			var (
				labels   map[string]types.NullString
				metadata v7action.Metadata
				org      v7action.Organization
				spaces   []v7action.Space
				domains  []v7action.Domain
			)

			BeforeEach(func() {
				labels = map[string]types.NullString{
					"k1": types.NewNullString("v1"),
					"k2": types.NewNullString("v2"),
				}

				metadata = v7action.Metadata{
					Labels: labels,
				}

				org = v7action.Organization{
					Name:     "org-name",
					GUID:     "org-guid",
					Metadata: &metadata,
				}

				spaces = []v7action.Space{
					v7action.Space{
						Name: "space-name-1",
						GUID: "space-guid-1",
					},
					v7action.Space{
						Name: "space-name-2",
						GUID: "space-guid-2",
					},
				}

				domains = []v7action.Domain{
					v7action.Domain{
						Name:             "yodie.com",
						GUID:             "yoodie.com-guid",
						OrganizationGUID: org.GUID,
					},
				}

				fakePluginActor.GetOrganizationByNameReturns(org, nil, nil)
				fakePluginActor.GetOrganizationSpacesReturns(spaces, nil, nil)
				fakePluginActor.GetOrganizationDomainsReturns(domains, nil, nil)
			})

			It("retrives the organization", func() {
				result := plugin_models.OrgSummary{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetOrganizationByNameCallCount()).To(Equal(1))
				orgName := fakePluginActor.GetOrganizationByNameArgsForCall(0)
				Expect(orgName).To(Equal(org.Name))
			})

			It("retrives the spaces for the organization", func() {
				result := plugin_models.OrgSummary{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetOrganizationSpacesCallCount()).To(Equal(1))
				orgGUID := fakePluginActor.GetOrganizationSpacesArgsForCall(0)
				Expect(orgGUID).To(Equal(org.GUID))
			})

			It("retrives the domains for the organization", func() {
				result := plugin_models.OrgSummary{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetOrganizationDomainsCallCount()).To(Equal(1))
				orgGUID, labelSelector := fakePluginActor.GetOrganizationDomainsArgsForCall(0)
				Expect(orgGUID).To(Equal(org.GUID))
				Expect(labelSelector).To(Equal(""))
			})

			It("populates the plugin model with the retrieved org, space, and domain information", func() {
				result := plugin_models.OrgSummary{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Name).To(Equal(org.Name))
				Expect(result.GUID).To(Equal(org.GUID))

				Expect(len(result.Spaces)).To(Equal(2))
				Expect(result.Spaces[1].Name).To(Equal(spaces[1].Name))

				Expect(len(result.Domains)).To(Equal(1))
				Expect(result.Domains[0].Name).To(Equal(domains[0].Name))
			})

			It("populates the plugin model with Metadata", func() {
				result := plugin_models.OrgSummary{}
				err := client.Call("CliRpcCmd.GetOrg", "org-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Metadata).ToNot(BeNil())
				Expect(result.Metadata.Labels).To(BeEquivalentTo(labels))
			})

			Context("when retrieving the org fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetOrganizationByNameReturns(v7action.Organization{}, nil, errors.New("org-error"))
				})

				It("returns an error", func() {
					result := plugin_models.OrgSummary{}
					err := client.Call("CliRpcCmd.GetOrg", "some-org", &result)
					Expect(err).To(MatchError("org-error"))
				})
			})

			Context("when retrieving the space fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetOrganizationSpacesReturns([]v7action.Space{}, nil, errors.New("space-error"))
				})

				It("returns an error", func() {
					result := plugin_models.OrgSummary{}
					err := client.Call("CliRpcCmd.GetOrg", "some-org", &result)
					Expect(err).To(MatchError("space-error"))
				})
			})
		})

		Describe("GetSpace", func() {
			var (
				labels   map[string]types.NullString
				metadata ccv3.Metadata
				space    v7action.Space
			)

			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "im-an-org-name",
					GUID: "and-im-an-org-guid",
				})

				labels = map[string]types.NullString{
					"k1": types.NewNullString("v1"),
					"k2": types.NewNullString("v2"),
				}

				metadata = ccv3.Metadata{
					Labels: labels,
				}

				space = v7action.Space{
					Name:     "space-name",
					GUID:     "space-guid",
					Metadata: &metadata,
				}

				fakePluginActor.GetSpaceByNameAndOrganizationReturns(space, v7action.Warnings{}, nil)
			})

			It("retrieves the space", func() {
				result := plugin_models.Space{}
				err := client.Call("CliRpcCmd.GetSpace", "space-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
				actualSpaceName, actualOrgGUID := fakePluginActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(actualSpaceName).To(Equal(space.Name))
				Expect(actualOrgGUID).To(Equal("and-im-an-org-guid"))
			})

			It("populates the plugin model with the retrieved space information", func() {
				result := plugin_models.Space{}
				err := client.Call("CliRpcCmd.GetSpace", "space-name", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(result.Name).To(Equal(space.Name))
				Expect(result.GUID).To(Equal(space.GUID))
				Expect(result.Metadata).ToNot(BeNil())
				Expect(result.Metadata.Labels).To(BeEquivalentTo(labels))
			})

			Context("when retrieving the space fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetSpaceByNameAndOrganizationReturns(v7action.Space{}, v7action.Warnings{}, errors.New("space-error"))
				})

				It("returns an error", func() {
					result := plugin_models.Space{}
					err := client.Call("CliRpcCmd.GetSpace", "space-name", &result)
					Expect(err).To(MatchError("space-error"))
				})
			})

			Context("when no org is targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						Name: "",
						GUID: "",
					})
				})
				It("complains that no org is targeted", func() {
					result := plugin_models.Space{}
					err := client.Call("CliRpcCmd.GetSpace", "space-name", &result)
					Expect(err).To(MatchError("no organization targeted"))
				})
			})
		})

		Describe("GetSpaces", func() {
			var (
				space1    v7action.Space
				space2    v7action.Space
				spaces    []v7action.Space
				labels1   map[string]types.NullString
				metadata1 ccv3.Metadata
				labels2   map[string]types.NullString
				metadata2 ccv3.Metadata
			)

			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "im-an-org-name",
					GUID: "and-im-an-org-guid",
				})

				labels1 = map[string]types.NullString{
					"k1": types.NewNullString("v1"),
					"k2": types.NewNullString("v2"),
				}

				metadata1 = ccv3.Metadata{
					Labels: labels1,
				}

				space1 = v7action.Space{
					Name:     "space1-name",
					GUID:     "space1-guid",
					Metadata: &metadata1,
				}

				labels2 = map[string]types.NullString{
					"b1": types.NewNullString("c1"),
					"b2": types.NewNullString("c2"),
				}

				metadata2 = ccv3.Metadata{
					Labels: labels2,
				}

				space2 = v7action.Space{
					Name:     "space2-name",
					GUID:     "space2-guid",
					Metadata: &metadata2,
				}
				spaces = []v7action.Space{space1, space2}

				fakePluginActor.GetOrganizationSpacesReturns(spaces, v7action.Warnings{}, nil)
			})

			It("retrieves the spaces", func() {
				result := []plugin_models.Space{}
				err := client.Call("CliRpcCmd.GetSpaces", "", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetOrganizationSpacesCallCount()).To(Equal(1))
				actualOrgGUID := fakePluginActor.GetOrganizationSpacesArgsForCall(0)
				Expect(actualOrgGUID).To(Equal("and-im-an-org-guid"))
			})

			It("populates the plugin model with the retrieved space information", func() {
				result := []plugin_models.Space{}
				err := client.Call("CliRpcCmd.GetSpaces", "", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(result)).To(Equal(2))
				Expect(result[0].Name).To(Equal(space1.Name))
				Expect(result[0].GUID).To(Equal(space1.GUID))
				Expect(result[0].Metadata).ToNot(BeNil())
				Expect(result[0].Metadata.Labels).To(BeEquivalentTo(labels1))
				Expect(result[1].Name).To(Equal(space2.Name))
				Expect(result[1].GUID).To(Equal(space2.GUID))
				Expect(result[1].Metadata).ToNot(BeNil())
				Expect(result[1].Metadata.Labels).To(BeEquivalentTo(labels2))
			})

			Context("when retrieving the spaces fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetOrganizationSpacesReturns([]v7action.Space{}, v7action.Warnings{}, errors.New("spaces-error"))
				})

				It("returns an error", func() {
					result := []plugin_models.Space{}
					err := client.Call("CliRpcCmd.GetSpaces", "", &result)
					Expect(err).To(MatchError("spaces-error"))
				})
			})

			Context("when no org is targeted", func() {
				BeforeEach(func() {
					fakeConfig.TargetedOrganizationReturns(configv3.Organization{
						Name: "",
						GUID: "",
					})
				})
				It("complains that no org is targeted", func() {
					result := []plugin_models.Space{}
					err := client.Call("CliRpcCmd.GetSpaces", "", &result)
					Expect(err).To(MatchError("no organization targeted"))
				})
			})
		})

		Describe("GetCurrentSpace", func() {
			BeforeEach(func() {
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					Name: "the-charlatans",
					GUID: "united-travel-service",
				})
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "the-actress",
					GUID: "family",
				})
				expectedSpace := v7action.Space{
					GUID: "united-travel-service",
					Name: "the-charlatans",
				}
				fakePluginActor.GetSpaceByNameAndOrganizationReturns(expectedSpace, v7action.Warnings{}, nil)
			})

			It("populates the plugin Space object with the current space settings in config", func() {
				result := plugin_models.CurrentSpace{}
				err := client.Call("CliRpcCmd.GetCurrentSpace", "", &result)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakePluginActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
				spaceName, orgGUID := fakePluginActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(spaceName).To(Equal("the-charlatans"))
				Expect(orgGUID).To(Equal("family"))

				Expect(result.Name).To(Equal("the-charlatans"))
				Expect(result.GUID).To(Equal("united-travel-service"))
			})

			Context("when retrieving the current space fails", func() {
				BeforeEach(func() {
					fakePluginActor.GetSpaceByNameAndOrganizationReturns(v7action.Space{}, nil, errors.New("some-error"))
				})

				It("returns an error", func() {
					result := plugin_models.CurrentSpace{}
					err := client.Call("CliRpcCmd.GetCurrentSpace", "", &result)
					Expect(err).To(MatchError("some-error"))
				})
			})
		})

		Describe("Username", func() {
			When("logged in", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("Yodie", nil)
				})
				It("returns the logged in username", func() {
					result := ""
					err := client.Call("CliRpcCmd.Username", "", &result)
					Expect(err).To(BeNil())
					Expect(result).To(Equal("Yodie"))
				})
			})

			When("not logged in", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("", nil)
				})
				It("returns the logged in username", func() {
					result := ""
					err := client.Call("CliRpcCmd.Username", "", &result)
					Expect(result).To(Equal(""))
					Expect(err).To(MatchError("not logged in"))
				})
			})
			When("config errors", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserNameReturns("", errors.New("config failed.."))
				})
				It("returns error", func() {
					result := ""
					err := client.Call("CliRpcCmd.Username", "", &result)
					Expect(result).To(Equal(""))
					Expect(err).To(MatchError("error processing config: config failed.."))
				})
			})
		})
		Describe("AccessToken", func() {

			BeforeEach(func() {
				fakePluginActor.RefreshAccessTokenReturns("token example", nil)
			})
			It("retrieves the access token", func() {

				result := ""
				err := client.Call("CliRpcCmd.AccessToken", "", &result)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakePluginActor.RefreshAccessTokenCallCount()).To(Equal(1))
				Expect(result).To(Equal("token example"))

			})
		})

		Describe("IsSkipSSLValidation", func() {
			When("skip ssl validation is false", func() {
				BeforeEach(func() {
					fakeConfig.SkipSSLValidationReturns(false)
				})

				It("returns false", func() {
					var result bool
					err = client.Call("CliRpcCmd.IsSkipSSLValidation", "", &result)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeFalse())
				})
			})
			When("skip ssl validation is true", func() {
				BeforeEach(func() {
					fakeConfig.SkipSSLValidationReturns(true)
				})

				It("returns false", func() {
					var result bool
					err = client.Call("CliRpcCmd.IsSkipSSLValidation", "", &result)

					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(BeTrue())
				})
			})
		})
	})

})

func pingCli(port string) {
	var connErr error
	var conn net.Conn
	for i := 0; i < 5; i++ {
		conn, connErr = net.Dial("tcp", "127.0.0.1:"+port)
		if connErr != nil {
			time.Sleep(200 * time.Millisecond)
		} else {
			conn.Close()
			break
		}
	}
	Expect(connErr).ToNot(HaveOccurred())
}
