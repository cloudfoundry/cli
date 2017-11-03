package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("deleted-orphaned-routes Command", func() {
	var (
		cmd             v2.DeleteOrphanedRoutesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeDeleteOrphanedRoutesActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeDeleteOrphanedRoutesActor)

		cmd = v2.DeleteOrphanedRoutesCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		Context("when checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: "faceman"}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
		})

		Context("when the user is logged in, and org and space are targeted", func() {
			BeforeEach(func() {
				fakeConfig.HasTargetedOrganizationReturns(true)
				fakeConfig.HasTargetedSpaceReturns(true)
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					GUID: "some-space-guid",
					Name: "some-space",
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

			Context("when getting the current user does not return an error", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(
						configv3.User{Name: "some-user"},
						nil)
				})

				Context("when the '-f' flag is provided", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					It("does not prompt for user confirmation", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).ToNot(Say("Really delete orphaned routes\\? \\[yN\\]:"))
					})
				})

				Context("when the '-f' flag is not provided", func() {
					Context("when user is prompted for confirmation", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("\n"))
							Expect(err).NotTo(HaveOccurred())
						})

						It("displays the interactive prompt", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Really delete orphaned routes\\? \\[yN\\]:"))
						})
					})

					Context("when the user inputs no", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("n\n"))
							Expect(err).NotTo(HaveOccurred())
						})

						It("does not delete orphaned routes", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeActor.GetOrphanedRoutesBySpaceCallCount()).To(Equal(0))
							Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
						})
					})

					Context("when the user input is invalid", func() {
						BeforeEach(func() {
							_, err := input.Write([]byte("e\n"))
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							Expect(executeErr).To(HaveOccurred())

							Expect(fakeActor.GetOrphanedRoutesBySpaceCallCount()).To(Equal(0))
							Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
						})
					})

					Context("when the user inputs yes", func() {
						var routes []v2action.Route

						BeforeEach(func() {
							_, err := input.Write([]byte("y\n"))
							Expect(err).NotTo(HaveOccurred())

							routes = []v2action.Route{
								{
									GUID: "route-1-guid",
									Host: "route-1",
									Domain: v2action.Domain{
										Name: "bosh-lite.com",
									},
									Path: "/path",
								},
								{
									GUID: "route-2-guid",
									Host: "route-2",
									Domain: v2action.Domain{
										Name: "bosh-lite.com",
									},
								},
							}

							fakeActor.GetOrphanedRoutesBySpaceReturns(routes, nil, nil)
						})

						It("displays getting routes message", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Getting routes as some-user ...\n"))
						})

						It("deletes the routes and displays that they are deleted", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(fakeActor.GetOrphanedRoutesBySpaceCallCount()).To(Equal(1))
							Expect(fakeActor.GetOrphanedRoutesBySpaceArgsForCall(0)).To(Equal("some-space-guid"))
							Expect(fakeActor.DeleteRouteCallCount()).To(Equal(2))
							Expect(fakeActor.DeleteRouteArgsForCall(0)).To(Equal(routes[0].GUID))
							Expect(fakeActor.DeleteRouteArgsForCall(1)).To(Equal(routes[1].GUID))

							Expect(testUI.Out).To(Say("Deleting route route-1.bosh-lite.com/path..."))
							Expect(testUI.Out).To(Say("Deleting route route-2.bosh-lite.com..."))
							Expect(testUI.Out).To(Say("OK"))
						})

						Context("when there are warnings", func() {
							BeforeEach(func() {
								fakeActor.GetOrphanedRoutesBySpaceReturns(
									[]v2action.Route{{GUID: "some-route-guid"}},
									[]string{"foo", "bar"},
									nil)
								fakeActor.DeleteRouteReturns([]string{"baz"}, nil)
							})

							It("displays the warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Err).To(Say("foo"))
								Expect(testUI.Err).To(Say("bar"))
								Expect(testUI.Err).To(Say("baz"))
							})
						})

						Context("when getting the routes returns an error", func() {
							var expectedErr error

							Context("when the error is a DomainNotFoundError", func() {
								BeforeEach(func() {
									fakeActor.GetOrphanedRoutesBySpaceReturns(
										nil,
										nil,
										actionerror.DomainNotFoundError{
											Name: "some-domain",
											GUID: "some-domain-guid",
										},
									)
								})

								It("returns translatableerror.DomainNotFoundError", func() {
									Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{
										Name: "some-domain",
										GUID: "some-domain-guid",
									}))
								})
							})

							Context("when the error is an OrphanedRoutesNotFoundError", func() {
								BeforeEach(func() {
									expectedErr = actionerror.OrphanedRoutesNotFoundError{}
									fakeActor.GetOrphanedRoutesBySpaceReturns(nil, nil, expectedErr)
								})

								It("should not return an error and only display 'OK'", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(fakeActor.DeleteRouteCallCount()).To(Equal(0))
								})
							})

							Context("when there is a generic error", func() {
								BeforeEach(func() {
									expectedErr = errors.New("getting orphaned routes error")
									fakeActor.GetOrphanedRoutesBySpaceReturns(nil, nil, expectedErr)
								})

								It("returns the error", func() {
									Expect(executeErr).To(MatchError(expectedErr))
								})
							})
						})

						Context("when deleting a route returns an error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("deleting route error")
								fakeActor.GetOrphanedRoutesBySpaceReturns(
									[]v2action.Route{{GUID: "some-route-guid"}},
									nil,
									nil)
								fakeActor.DeleteRouteReturns(nil, expectedErr)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError(expectedErr))
							})
						})
					})
				})
			})
		})
	})
})
