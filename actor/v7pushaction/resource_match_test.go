package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MatchResources", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		resources  []sharedaction.V3Resource
		executeErr error

		matched   []sharedaction.V3Resource
		unmatched []sharedaction.V3Resource
		warnings  Warnings
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, _ = getTestPushActor()
	})

	JustBeforeEach(func() {
		matched, unmatched, warnings, executeErr = actor.MatchResources(resources)
	})

	When("all the resources are unmatched", func() {
		BeforeEach(func() {
			resources = []sharedaction.V3Resource{
				{Checksum: ccv3.Checksum{Value: "file-1"}},
				{Checksum: ccv3.Checksum{Value: "file-2"}},
				{Checksum: ccv3.Checksum{Value: "file-3"}},
			}

			fakeV7Actor.ResourceMatchReturns(
				[]sharedaction.V3Resource{},
				v7action.Warnings{"all-unmatched-warning"},
				nil,
			)
		})

		It("returns an empty slice of matches and a complete slice of unmatches (in the order provided)", func() {
			Expect(executeErr).To(BeNil())
			Expect(matched).To(BeEmpty())
			Expect(unmatched).To(Equal(resources))
			Expect(warnings).To(Equal(Warnings{"all-unmatched-warning"}))
		})
	})

	When("all the resources are matched", func() {
		BeforeEach(func() {
			resources = []sharedaction.V3Resource{
				{Checksum: ccv3.Checksum{Value: "file-1"}},
				{Checksum: ccv3.Checksum{Value: "file-2"}},
				{Checksum: ccv3.Checksum{Value: "file-3"}},
			}

			fakeV7Actor.ResourceMatchReturns(
				resources,
				v7action.Warnings{"all-unmatched-warning"},
				nil,
			)
		})

		It("returns a complete slice of matches and an empty slice of unmatches (in the order provided)", func() {
			Expect(executeErr).To(BeNil())
			Expect(matched).To(Equal(resources))
			Expect(unmatched).To(BeEmpty())
			Expect(warnings).To(Equal(Warnings{"all-unmatched-warning"}))
		})
	})

	When("some of the resources are matched", func() {
		var expectedMatches []sharedaction.V3Resource
		var expectedUnmatches []sharedaction.V3Resource

		BeforeEach(func() {
			resources = []sharedaction.V3Resource{
				{Checksum: ccv3.Checksum{Value: "file-1"}},
				{Checksum: ccv3.Checksum{Value: "file-2"}},
				{Checksum: ccv3.Checksum{Value: "file-3"}},
			}

			expectedMatches = []sharedaction.V3Resource{
				{Checksum: ccv3.Checksum{Value: "file-1"}},
				{Checksum: ccv3.Checksum{Value: "file-3"}},
			}

			expectedUnmatches = []sharedaction.V3Resource{
				{Checksum: ccv3.Checksum{Value: "file-2"}},
			}

			fakeV7Actor.ResourceMatchReturns(
				expectedMatches,
				v7action.Warnings{"all-unmatched-warning"},
				nil,
			)
		})

		It("returns a slice of matches and a slice of unmatches (in the order provided)", func() {
			Expect(executeErr).To(BeNil())
			Expect(matched).To(Equal(expectedMatches))
			Expect(unmatched).To(Equal(expectedUnmatches))
			Expect(warnings).To(Equal(Warnings{"all-unmatched-warning"}))
		})
	})

	When("v7actor.ResourceMatch returns an error", func() {
		BeforeEach(func() {
			fakeV7Actor.ResourceMatchReturns(
				nil,
				v7action.Warnings{
					"error-response-warning-1",
					"error-response-warning-2",
				},
				errors.New("error-response-text"),
			)
		})

		It("returns the same error", func() {
			Expect(executeErr).To(MatchError("error-response-text"))
			Expect(warnings).To(ConsistOf(
				"error-response-warning-1",
				"error-response-warning-2",
			))
		})
	})
})
