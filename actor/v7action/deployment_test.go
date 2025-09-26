package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployment Actions", func() {
	var (
		actor                     *Actor
		executeErr                error
		warnings                  Warnings
		returnedDeploymentGUID    string
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
		fakeCloudControllerClient.CreateApplicationDeploymentReturns(
			"some-deployment-guid",
			ccv3.Warnings{"create-warning-1", "create-warning-2"},
			errors.New("create-error"),
		)
	})

	Describe("Create a deployment with app and revision guids", func() {
		JustBeforeEach(func() {
			var dep resources.Deployment
			dep.Strategy = constant.DeploymentStrategyRolling
			dep.RevisionGUID = "some-revision-guid"
			dep.Relationships = resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"}}
			returnedDeploymentGUID, warnings, executeErr = actor.CreateDeployment(dep)
		})

		When("the client fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateApplicationDeploymentReturns(
					"some-deployment-guid",
					ccv3.Warnings{"create-warning-1", "create-warning-2"},
					errors.New("create-deployment-error"),
				)
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError("create-deployment-error"))
				Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2"))
			})
		})

		It("delegates to the cloud controller client", func() {

			Expect(fakeCloudControllerClient.CreateApplicationDeploymentCallCount()).To(Equal(1), "CreateApplicationDeploymentByRevision call count")
			dep := fakeCloudControllerClient.CreateApplicationDeploymentArgsForCall(0)

			Expect(dep.Relationships[constant.RelationshipTypeApplication].GUID).To(Equal("some-app-guid"))
			Expect(dep.RevisionGUID).To(Equal("some-revision-guid"))

			Expect(returnedDeploymentGUID).To(Equal("some-deployment-guid"))
			Expect(warnings).To(Equal(Warnings{"create-warning-1", "create-warning-2"}))
		})
	})

	Describe("Create a deployment with app and droplet guids", func() {
		It("delegates to the cloud controller client", func() {
			fakeCloudControllerClient.CreateApplicationDeploymentReturns("some-deployment-guid", ccv3.Warnings{"create-warning-1", "create-warning-2"}, errors.New("create-error"))

			var dep resources.Deployment
			dep.Strategy = constant.DeploymentStrategyCanary
			dep.Relationships = resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"}}
			dep.DropletGUID = "some-droplet-guid"
			returnedDeploymentGUID, warnings, executeErr := actor.CreateDeployment(dep)

			Expect(fakeCloudControllerClient.CreateApplicationDeploymentCallCount()).To(Equal(1))
			dep1 := fakeCloudControllerClient.CreateApplicationDeploymentArgsForCall(0)

			Expect(dep1.Relationships[constant.RelationshipTypeApplication].GUID).To(Equal("some-app-guid"))
			Expect(dep1.DropletGUID).To(Equal("some-droplet-guid"))

			Expect(returnedDeploymentGUID).To(Equal("some-deployment-guid"))
			Expect(warnings).To(Equal(Warnings{"create-warning-1", "create-warning-2"}))
			Expect(executeErr).To(MatchError("create-error"))
		})
	})

	Describe("GetLatestActiveDeploymentForApp", func() {
		var (
			executeErr error
			warnings   Warnings
			deployment resources.Deployment

			appGUID string
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			deployment, warnings, executeErr = actor.GetLatestActiveDeploymentForApp(appGUID)
		})

		It("delegates to the CC client", func() {
			Expect(fakeCloudControllerClient.GetDeploymentsCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.GetDeploymentsArgsForCall(0)).To(Equal(
				[]ccv3.Query{
					{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
					{Key: ccv3.StatusValueFilter, Values: []string{string(constant.DeploymentStatusValueActive)}},
					{Key: ccv3.OrderBy, Values: []string{ccv3.CreatedAtDescendingOrder}},
					{Key: ccv3.PerPage, Values: []string{"1"}},
					{Key: ccv3.Page, Values: []string{"1"}},
				},
			))
		})

		When("the cc client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentsReturns(
					[]resources.Deployment{},
					ccv3.Warnings{"get-deployments-warning"},
					errors.New("get-deployments-error"),
				)
			})

			It("returns an error and warnings", func() {
				Expect(executeErr).To(MatchError("get-deployments-error"))
				Expect(warnings).To(ConsistOf("get-deployments-warning"))
			})
		})

		When("there are no deployments returned", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentsReturns(
					[]resources.Deployment{},
					ccv3.Warnings{"get-deployments-warning"},
					nil,
				)
			})

			It("returns a deployment not found error and warnings", func() {
				Expect(executeErr).To(Equal(actionerror.ActiveDeploymentNotFoundError{}))
				Expect(warnings).To(ConsistOf("get-deployments-warning"))
			})

		})

		When("everything succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDeploymentsReturns(
					[]resources.Deployment{{GUID: "dep-guid"}},
					ccv3.Warnings{"get-deployments-warning"},
					nil,
				)
			})

			It("returns a deployment not found error and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-deployments-warning"))
				Expect(deployment).To(Equal(resources.Deployment{GUID: "dep-guid"}))
			})

		})
	})

	Describe("CancelDeployment", func() {
		var (
			deploymentGUID string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			deploymentGUID = "dep-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CancelDeployment(deploymentGUID)
		})

		It("delegates to the cc client", func() {
			Expect(fakeCloudControllerClient.CancelDeploymentCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.CancelDeploymentArgsForCall(0)).To(Equal(deploymentGUID))
		})

		When("the client fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CancelDeploymentReturns(ccv3.Warnings{"cancel-deployment-warnings"}, errors.New("cancel-deployment-error"))
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError("cancel-deployment-error"))
				Expect(warnings).To(ConsistOf("cancel-deployment-warnings"))
			})

		})

		When("the client succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CancelDeploymentReturns(ccv3.Warnings{"cancel-deployment-warnings"}, nil)
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("cancel-deployment-warnings"))
			})
		})
	})

	Describe("ContinueDeployment", func() {
		var (
			deploymentGUID string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			deploymentGUID = "dep-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.ContinueDeployment(deploymentGUID)
		})

		It("delegates to the cc client", func() {
			Expect(fakeCloudControllerClient.ContinueDeploymentCallCount()).To(Equal(1))
			Expect(fakeCloudControllerClient.ContinueDeploymentArgsForCall(0)).To(Equal(deploymentGUID))
		})

		When("the client fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.ContinueDeploymentReturns(ccv3.Warnings{"continue-deployment-warnings"}, errors.New("continue-deployment-error"))
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError("continue-deployment-error"))
				Expect(warnings).To(ConsistOf("continue-deployment-warnings"))
			})

		})

		When("the client succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.ContinueDeploymentReturns(ccv3.Warnings{"continue-deployment-warnings"}, nil)
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("continue-deployment-warnings"))
			})
		})
	})
})
