package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Instance Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("ApplicationInstance", func() {
		var instance ApplicationInstance

		BeforeEach(func() {
			instance = ApplicationInstance{}
		})

		Describe("Crashed", func() {
			Context("instance is crashed", func() {
				It("returns true", func() {
					instance.State = constant.ApplicationInstanceCrashed
					Expect(instance.Crashed()).To(BeTrue())
				})
			})

			Context("instance is *not* crashed", func() {
				It("returns false", func() {
					instance.State = constant.ApplicationInstanceRunning
					Expect(instance.Crashed()).To(BeFalse())
				})
			})
		})

		Describe("Flapping", func() {
			Context("instance is flapping", func() {
				It("returns true", func() {
					instance.State = constant.ApplicationInstanceFlapping
					Expect(instance.Flapping()).To(BeTrue())
				})
			})

			Context("instance is *not* flapping", func() {
				It("returns false", func() {
					instance.State = constant.ApplicationInstanceCrashed
					Expect(instance.Flapping()).To(BeFalse())
				})
			})
		})

		Describe("Running", func() {
			Context("instance is running", func() {
				It("returns true", func() {
					instance.State = constant.ApplicationInstanceRunning
					Expect(instance.Running()).To(BeTrue())
				})
			})

			Context("instance is *not* running", func() {
				It("returns false", func() {
					instance.State = constant.ApplicationInstanceCrashed
					Expect(instance.Running()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetApplicationInstancesByApplication", func() {
		When("the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
					map[int]ccv2.ApplicationInstance{
						0: {ID: 0, Details: "hello", Since: 1485985587.12345, State: constant.ApplicationInstanceRunning},
						1: {ID: 1, Details: "hi", Since: 1485985587.567},
					},
					ccv2.Warnings{"instance-warning-1", "instance-warning-2"},
					nil)
			})

			It("returns the application instances and all warnings", func() {
				instances, warnings, err := actor.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(instances).To(ConsistOf(
					ApplicationInstance{
						ID:      0,
						Details: "hello",
						Since:   1485985587.12345,
						State:   constant.ApplicationInstanceRunning,
					},
					ApplicationInstance{
						ID:      1,
						Details: "hi",
						Since:   1485985587.567,
					},
				))
				Expect(warnings).To(ConsistOf("instance-warning-1", "instance-warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationApplicationInstancesArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		When("an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
					nil,
					ccv2.Warnings{"instances-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("instances-warning"))
			})

			When("the application does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
						nil,
						nil,
						ccerror.ResourceNotFoundError{})
				})

				It("returns an ApplicationInstancesNotFoundError", func() {
					_, _, err := actor.GetApplicationInstancesByApplication("some-app-guid")
					Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
				})
			})

			When("the app has not been staged yet", func() {
				When("getting instances returns a CF-NotStaged error", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
							nil,
							nil,
							ccerror.NotStagedError{})
					})

					It("returns an ApplicationInstancesNotFoundError", func() {
						_, _, err := actor.GetApplicationInstancesByApplication("some-app-guid")
						Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
					})
				})
			})

			When("getting instances returns a CF-InstancesError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationApplicationInstancesReturns(
						nil,
						nil,
						ccerror.InstancesError{})
				})

				It("returns an ApplicationInstancesNotFoundError", func() {
					_, _, err := actor.GetApplicationInstancesByApplication("some-app-guid")
					Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
				})
			})
		})
	})
})
