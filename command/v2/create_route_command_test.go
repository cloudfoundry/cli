package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Create Route Command", func() {
	var (
		cmd             CreateRouteCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeCreateRouteActor
		binaryName      string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeCreateRouteActor)

		cmd = CreateRouteCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		cmd.RequiredArgs.Space = "some-space"
		cmd.RequiredArgs.Domain = "some-domain"

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeActor.CloudControllerAPIVersionReturns(ccversion.MinVersionTCPRouting)
	})

	DescribeTable("argument combinations",
		func(expectedErr error, hostname string, path string, port flag.Port, randomPort bool) {
			cmd.Port = port
			cmd.Hostname = hostname
			cmd.Path = path
			cmd.RandomPort = randomPort

			executeErr := cmd.Execute(nil)
			if expectedErr == nil {
				Expect(executeErr).To(BeNil())
			} else {
				Expect(executeErr).To(Equal(expectedErr))
			}
		},
		Entry("hostname", nil, "some-hostname", "", flag.Port{NullInt: types.NullInt{IsSet: false}}, false),
		Entry("path", nil, "", "some-path", flag.Port{NullInt: types.NullInt{IsSet: false}}, false),
		Entry("hostname and path", nil, "some-hostname", "some-path", flag.Port{NullInt: types.NullInt{IsSet: false}}, false),
		Entry("hostname and port", translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "--port"}}, "some-hostname", "", flag.Port{NullInt: types.NullInt{IsSet: true}}, false),
		Entry("path and port", translatableerror.ArgumentCombinationError{Args: []string{"--path", "--port"}}, "", "some-path", flag.Port{NullInt: types.NullInt{IsSet: true}}, false),
		Entry("hostname, path, and port", translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "--path", "--port"}}, "some-hostname", "some-path", flag.Port{NullInt: types.NullInt{IsSet: true}}, false),
		Entry("hostname and random port", translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "--random-port"}}, "some-hostname", "", flag.Port{NullInt: types.NullInt{IsSet: false}}, true),
		Entry("path and random port", translatableerror.ArgumentCombinationError{Args: []string{"--path", "--random-port"}}, "", "some-path", flag.Port{NullInt: types.NullInt{IsSet: false}}, true),
		Entry("hostname, path, and random port", translatableerror.ArgumentCombinationError{Args: []string{"--hostname", "--path", "--random-port"}}, "some-hostname", "some-path", flag.Port{NullInt: types.NullInt{IsSet: false}}, true),
		Entry("port", nil, "", "", flag.Port{NullInt: types.NullInt{IsSet: true}}, false),
		Entry("random port", nil, "", "", flag.Port{NullInt: types.NullInt{IsSet: false}}, true),
		Entry("port and random port", translatableerror.ArgumentCombinationError{Args: []string{"--port", "--random-port"}}, "", "", flag.Port{NullInt: types.NullInt{IsSet: true}}, true),
	)

	DescribeTable("minimum api version checks",
		func(expectedErr error, port flag.Port, randomPort bool, path string, apiVersion string) {
			cmd.Port = port
			cmd.RandomPort = randomPort
			cmd.Path = path
			fakeActor.CloudControllerAPIVersionReturns(apiVersion)

			executeErr := cmd.Execute(nil)
			if expectedErr == nil {
				Expect(executeErr).To(BeNil())
			} else {
				Expect(executeErr).To(Equal(expectedErr))
			}
		},

		Entry("port, CC Version 2.52.0", translatableerror.MinimumAPIVersionNotMetError{
			Command:        "Option '--port'",
			CurrentVersion: "2.52.0",
			MinimumVersion: ccversion.MinVersionTCPRouting,
		}, flag.Port{NullInt: types.NullInt{IsSet: true}}, false, "", "2.52.0"),

		Entry("port, CC Version 2.53.0", nil, flag.Port{NullInt: types.NullInt{IsSet: true}}, false, "", ccversion.MinVersionTCPRouting),

		Entry("random-port, CC Version 2.52.0", translatableerror.MinimumAPIVersionNotMetError{
			Command:        "Option '--random-port'",
			CurrentVersion: "2.52.0",
			MinimumVersion: ccversion.MinVersionTCPRouting,
		}, flag.Port{}, true, "", "2.52.0"),

		Entry("random-port, CC Version 2.53.0", nil, flag.Port{}, true, "", ccversion.MinVersionTCPRouting),

		Entry("path, CC Version 2.35.0", translatableerror.MinimumAPIVersionNotMetError{
			Command:        "Option '--path'",
			CurrentVersion: "2.35.0",
			MinimumVersion: ccversion.MinVersionHTTPRoutePath,
		}, flag.Port{}, false, "some-path", "2.35.0"),

		Entry("path, CC Version 2.36.0", nil, flag.Port{}, false, "some-path", ccversion.MinVersionHTTPRoutePath),
	)

	Context("when all the arguments check out", func() {
		var executeErr error

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(translatableerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error if the check fails", func() {
				Expect(executeErr).To(MatchError(translatableerror.NotLoggedInError{BinaryName: "faceman"}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeFalse())
			})
		})

		Context("when getting the current user returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("getting current user error")
				fakeConfig.CurrentUserReturns(
					configv3.User{},
					expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})

		Context("when the user is logged in, and the org is targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "some-org-guid", Name: "some-org"})
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
			})

			Context("when no flags are provided", func() {
				BeforeEach(func() {
					fakeActor.CreateRouteWithExistenceCheckReturns(v2action.Route{
						Domain: v2action.Domain{
							Name: "some-domain",
						}}, v2action.Warnings{"create-route-warning-1", "create-route-warning-2"}, nil)
				})

				It("creates a route with existence check", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Creating route some-domain for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Out).To(Say("Route some-domain has been created\\."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
					orgGUID, spaceName, route, generatePort := fakeActor.CreateRouteWithExistenceCheckArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(route.Host).To(BeEmpty())
					Expect(route.Path).To(BeEmpty())
					Expect(route.Port).To(Equal(types.NullInt{IsSet: false}))
					Expect(generatePort).To(BeFalse())
				})
			})

			Context("when host and path flags are provided", func() {
				BeforeEach(func() {
					cmd.Hostname = "some-host"
					cmd.Path = "some-path"

					fakeActor.CreateRouteWithExistenceCheckReturns(v2action.Route{
						Domain: v2action.Domain{
							Name: "some-domain",
						},
						Host: "some-host",
						Path: "some-path",
					}, v2action.Warnings{"create-route-warning-1", "create-route-warning-2"}, nil)
				})

				It("creates a route with existence check", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Creating route some-host.some-domain/some-path for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Out).To(Say("Route some-host.some-domain/some-path has been created\\."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
					orgGUID, spaceName, route, generatePort := fakeActor.CreateRouteWithExistenceCheckArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(route.Host).To(Equal("some-host"))
					Expect(route.Path).To(Equal("some-path"))
					Expect(route.Port).To(Equal(types.NullInt{IsSet: false}))
					Expect(generatePort).To(BeFalse())
				})
			})

			Context("when port flag is provided", func() {
				BeforeEach(func() {
					cmd.Port = flag.Port{NullInt: types.NullInt{Value: 42, IsSet: true}}

					fakeActor.CreateRouteWithExistenceCheckReturns(v2action.Route{
						Domain: v2action.Domain{
							Name: "some-domain",
						},
						Port: types.NullInt{IsSet: true, Value: 42},
					}, v2action.Warnings{"create-route-warning-1", "create-route-warning-2"}, nil)
				})

				It("creates a route with existence check", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Creating route some-domain:42 for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Out).To(Say("Route some-domain:42 has been created\\."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
					orgGUID, spaceName, route, generatePort := fakeActor.CreateRouteWithExistenceCheckArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(route.Host).To(BeEmpty())
					Expect(route.Path).To(BeEmpty())
					Expect(route.Port).To(Equal(types.NullInt{IsSet: true, Value: 42}))
					Expect(generatePort).To(BeFalse())
				})
			})

			Context("when random-port flag is provided", func() {
				BeforeEach(func() {
					cmd.RandomPort = true
					fakeActor.CreateRouteWithExistenceCheckReturns(v2action.Route{
						Domain: v2action.Domain{
							Name: "some-domain",
						},
						Port: types.NullInt{IsSet: true, Value: 1115},
					}, v2action.Warnings{"create-route-warning-1", "create-route-warning-2"}, nil)
				})

				It("creates a route with existence check", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Creating route some-domain for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Out).To(Say("Route some-domain:1115 has been created\\."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
					orgGUID, spaceName, route, generatePort := fakeActor.CreateRouteWithExistenceCheckArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))
					Expect(route.Host).To(BeEmpty())
					Expect(route.Path).To(BeEmpty())
					Expect(route.Port).To(Equal(types.NullInt{IsSet: false}))
					Expect(generatePort).To(BeTrue())
				})
			})

			Context("when creating route returns a DomainNotFoundError error", func() {
				BeforeEach(func() {
					fakeActor.CreateRouteWithExistenceCheckReturns(
						v2action.Route{},
						v2action.Warnings{"create-route-warning-1", "create-route-warning-2"},
						actionerror.DomainNotFoundError{Name: "some-domain"},
					)
				})

				It("prints warnings and returns an error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{Name: "some-domain"}))

					Expect(testUI.Out).To(Say("Creating route some-domain for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Out).NotTo(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
				})
			})

			Context("when creating route returns a RouteAlreadyExistsError error", func() {
				BeforeEach(func() {
					cmd.Hostname = "some-host"

					fakeActor.CreateRouteWithExistenceCheckReturns(
						v2action.Route{},
						v2action.Warnings{"create-route-warning-1", "create-route-warning-2"},
						actionerror.RouteAlreadyExistsError{
							Route: v2action.Route{Host: "some-host"}.String(),
						},
					)
				})

				It("prints warnings and returns an error", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Out).To(Say("Creating route some-host\\.some-domain for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Err).To(Say("Route some-host\\.some-domain already exists\\."))
					Expect(testUI.Out).To(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
				})
			})

			Context("when creating route returns a generic error", func() {
				var createRouteErr error
				BeforeEach(func() {
					createRouteErr = errors.New("Oh nooes")
					fakeActor.CreateRouteWithExistenceCheckReturns(v2action.Route{}, v2action.Warnings{"create-route-warning-1", "create-route-warning-2"}, createRouteErr)
				})

				It("prints warnings and returns an error", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr).To(MatchError(createRouteErr))

					Expect(testUI.Out).To(Say("Creating route some-domain for org some-org / space some-space as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("create-route-warning-1"))
					Expect(testUI.Err).To(Say("create-route-warning-2"))
					Expect(testUI.Out).NotTo(Say("OK"))

					Expect(fakeActor.CreateRouteWithExistenceCheckCallCount()).To(Equal(1))
				})
			})
		})
	})
})
