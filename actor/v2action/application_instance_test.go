package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
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
					instance.State = ccv2.ApplicationInstanceCrashed
					Expect(instance.Crashed()).To(BeTrue())
				})
			})

			Context("instance is *not* crashed", func() {
				It("returns false", func() {
					instance.State = ccv2.ApplicationInstanceRunning
					Expect(instance.Crashed()).To(BeFalse())
				})
			})
		})

		Describe("Flapping", func() {
			Context("instance is flapping", func() {
				It("returns true", func() {
					instance.State = ccv2.ApplicationInstanceFlapping
					Expect(instance.Flapping()).To(BeTrue())
				})
			})

			Context("instance is *not* flapping", func() {
				It("returns false", func() {
					instance.State = ccv2.ApplicationInstanceCrashed
					Expect(instance.Flapping()).To(BeFalse())
				})
			})
		})

		Describe("Running", func() {
			Context("instance is running", func() {
				It("returns true", func() {
					instance.State = ccv2.ApplicationInstanceRunning
					Expect(instance.Running()).To(BeTrue())
				})
			})

			Context("instance is *not* running", func() {
				It("returns false", func() {
					instance.State = ccv2.ApplicationInstanceCrashed
					Expect(instance.Running()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetApplicationInstancesByApplication", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationInstancesByApplicationReturns(
					map[int]ccv2.ApplicationInstance{
						0: {ID: 0, Details: "hello", Since: 1485985587.12345, State: ccv2.ApplicationInstanceRunning},
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
						State:   ccv2.ApplicationInstanceRunning,
					},
					ApplicationInstance{
						ID:      1,
						Details: "hi",
						Since:   1485985587.567,
					},
				))
				Expect(warnings).To(ConsistOf("instance-warning-1", "instance-warning-2"))

				Expect(fakeCloudControllerClient.GetApplicationInstancesByApplicationCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationInstancesByApplicationArgsForCall(0)).To(Equal("some-app-guid"))
			})
		})

		Context("when an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("banana")
				fakeCloudControllerClient.GetApplicationInstancesByApplicationReturns(
					nil,
					ccv2.Warnings{"instances-warning"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				_, warnings, err := actor.GetApplicationInstancesByApplication("some-app-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("instances-warning"))
			})

			Context("when the application does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationInstancesByApplicationReturns(
						nil,
						nil,
						ccerror.ResourceNotFoundError{})
				})

				It("returns an ApplicationInstancesNotFoundError", func() {
					_, _, err := actor.GetApplicationInstancesByApplication("some-app-guid")
					Expect(err).To(MatchError(actionerror.ApplicationInstancesNotFoundError{ApplicationGUID: "some-app-guid"}))
				})
			})

			Context("when the app has not been staged yet", func() {
				Context("when getting instances returns a CF-NotStaged error", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationInstancesByApplicationReturns(
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

			Context("when getting instances returns a CF-InstancesError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationInstancesByApplicationReturns(
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
