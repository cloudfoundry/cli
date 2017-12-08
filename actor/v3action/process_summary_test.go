package v3action_test

import (
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Actions", func() {
	Describe("ProcessSummary", func() {
		var summary ProcessSummary
		BeforeEach(func() {
			summary = ProcessSummary{
				InstanceDetails: []Instance{
					Instance{State: constant.ProcessInstanceRunning},
					Instance{State: constant.ProcessInstanceRunning},
					Instance{State: constant.ProcessInstanceDown},
				},
			}
		})

		Describe("TotalInstanceCount", func() {
			It("returns the total number of instances", func() {
				Expect(summary.TotalInstanceCount()).To(Equal(3))
			})
		})

		Describe("HealthyInstanceCount", func() {
			It("returns the total number of RUNNING instances", func() {
				Expect(summary.HealthyInstanceCount()).To(Equal(2))
			})
		})
	})

	Describe("ProcessSummaries", func() {
		var summaries ProcessSummaries

		BeforeEach(func() {
			summaries = ProcessSummaries{
				{
					Process: Process{
						Type: "worker",
					},
					InstanceDetails: []Instance{
						{State: constant.ProcessInstanceRunning},
						{State: constant.ProcessInstanceDown},
					},
				},
				{
					Process: Process{
						Type: "console",
					},
					InstanceDetails: []Instance{
						{State: constant.ProcessInstanceRunning},
					},
				},
				{
					Process: Process{
						Type: constant.ProcessTypeWeb,
					},
					InstanceDetails: []Instance{
						{State: constant.ProcessInstanceRunning},
						{State: constant.ProcessInstanceRunning},
						{State: constant.ProcessInstanceDown},
					},
				},
			}
		})

		Describe("Sort", func() {
			It("sorts processes with web first and then alphabetically sorted", func() {
				summaries.Sort()
				Expect(summaries[0].Type).To(Equal(constant.ProcessTypeWeb))
				Expect(summaries[1].Type).To(Equal("console"))
				Expect(summaries[2].Type).To(Equal("worker"))
			})
		})

		Describe("String", func() {
			It("returns all processes and their instance count ", func() {
				Expect(summaries.String()).To(Equal("web:2/3, console:1/1, worker:1/2"))
			})
		})
	})
})
