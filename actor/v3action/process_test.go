package v3action_test

import (
	"time"

	. "code.cloudfoundry.org/cli/actor/v3action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Process Actions", func() {
	Describe("Instance", func() {
		Describe("StartTime", func() {
			It("returns the time that the instance started", func() {
				instance := Instance{Uptime: 86400}
				Expect(instance.StartTime()).To(BeTemporally("~", time.Now().Add(-24*time.Hour), 10*time.Second))
			})
		})
	})

	Describe("Process", func() {
		var process Process
		BeforeEach(func() {
			process = Process{
				Instances: []Instance{
					Instance{State: "RUNNING"},
					Instance{State: "RUNNING"},
					Instance{State: "STOPPED"},
				},
			}
		})

		Describe("TotalInstanceCount", func() {
			It("returns the total number of instances", func() {
				Expect(process.TotalInstanceCount()).To(Equal(3))
			})
		})

		Describe("HealthyInstanceCount", func() {
			It("returns the total number of RUNNING instances", func() {
				Expect(process.HealthyInstanceCount()).To(Equal(2))
			})
		})
	})

	Describe("Processes", func() {
		var processes Processes

		BeforeEach(func() {
			processes = Processes{
				{
					Type: "worker",
					Instances: []Instance{
						{State: "RUNNING"},
						{State: "STOPPED"},
					},
				},
				{
					Type: "console",
					Instances: []Instance{
						{State: "RUNNING"},
					},
				},
				{
					Type: "web",
					Instances: []Instance{
						{State: "RUNNING"},
						{State: "RUNNING"},
						{State: "STOPPED"},
					},
				},
			}
		})

		Describe("Sort", func() {
			It("sorts processes with web first and then alphabetically sorted", func() {
				processes.Sort()
				Expect(processes[0].Type).To(Equal("web"))
				Expect(processes[1].Type).To(Equal("console"))
				Expect(processes[2].Type).To(Equal("worker"))
			})
		})

		Describe("Summary", func() {
			It("returns all processes and their instance count ", func() {
				Expect(processes.Summary()).To(Equal("web:2/3, console:1/1, worker:1/2"))
			})
		})
	})
})
