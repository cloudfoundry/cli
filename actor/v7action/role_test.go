package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Role Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()
	})

	Describe("CreateOrgRole", func() {
		var (
			role       Role
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			role, warnings, executeErr = actor.CreateOrgRole(constant.OrgAuditorRole, "user-guid", "org-guid")
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
					ccv3.Warnings{"create-role-warning"},
					nil,
				)
			})

			It("returns the role and any warnings", func() {
				Expect(role).To(Equal(
					Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
				))
				Expect(warnings).To(ConsistOf("create-role-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				passedRole := fakeCloudControllerClient.CreateRoleArgsForCall(0)

				Expect(passedRole).To(Equal(
					ccv3.Role{
						Type:     constant.OrgAuditorRole,
						UserGUID: "user-guid",
						OrgGUID:  "org-guid",
					},
				))
			})
		})

		When("the API call to create the role returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRoleReturns(
					ccv3.Role{},
					ccv3.Warnings{"create-role-warning"},
					errors.New("create-role-error"),
				)
			})

			It("it returns an error and warnings", func() {
				Expect(fakeCloudControllerClient.CreateRoleCallCount()).To(Equal(1))
				Expect(warnings).To(ConsistOf("create-role-warning"))
				Expect(executeErr).To(MatchError("create-role-error"))
			})
		})
	})
})
