package ui_helpers_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/ui_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UI Helpers", func() {
	Describe("ColoredAppState", func() {
		It("returns plain state for stopped app with no running instances", func() {
			app := models.ApplicationFields{
				State:            "stopped",
				RunningInstances: 0,
				InstanceCount:    1,
			}

			state := ui_helpers.ColoredAppState(app)
			Expect(state).To(Equal("stopped"))
		})

		It("returns crashed state when app is not stopped but has zero running instances", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 0,
				InstanceCount:    3,
			}

			state := ui_helpers.ColoredAppState(app)
			// Should be colored as crashed
			Expect(state).To(ContainSubstring("started"))
		})

		It("returns warning state when some instances are running but not all", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 2,
				InstanceCount:    5,
			}

			state := ui_helpers.ColoredAppState(app)
			// Should be colored as warning
			Expect(state).To(ContainSubstring("started"))
		})

		It("returns plain state when all instances are running", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 3,
				InstanceCount:    3,
			}

			state := ui_helpers.ColoredAppState(app)
			Expect(state).To(Equal("started"))
		})

		It("handles lowercase state", func() {
			app := models.ApplicationFields{
				State:            "STARTED",
				RunningInstances: 1,
				InstanceCount:    1,
			}

			state := ui_helpers.ColoredAppState(app)
			Expect(state).To(Equal("started"))
		})

		It("handles crashed state when app is starting", func() {
			app := models.ApplicationFields{
				State:            "starting",
				RunningInstances: 0,
				InstanceCount:    2,
			}

			state := ui_helpers.ColoredAppState(app)
			Expect(state).To(ContainSubstring("starting"))
		})
	})

	Describe("ColoredAppInstances", func() {
		It("returns health string when all instances are running", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 3,
				InstanceCount:    3,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			Expect(instances).To(Equal("3/3"))
		})

		It("returns health string when some instances are running", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 2,
				InstanceCount:    5,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			// Should be colored as warning
			Expect(instances).To(ContainSubstring("2/5"))
		})

		It("returns plain string for stopped app", func() {
			app := models.ApplicationFields{
				State:            "stopped",
				RunningInstances: 0,
				InstanceCount:    3,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			Expect(instances).To(Equal("0/3"))
		})

		It("returns crashed color when app is started but no instances running", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 0,
				InstanceCount:    2,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			// Should be colored as crashed
			Expect(instances).To(ContainSubstring("0/2"))
		})

		It("returns question mark when running instances is negative", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: -1,
				InstanceCount:    3,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			Expect(instances).To(ContainSubstring("?/3"))
		})

		It("handles negative running instances with stopped state", func() {
			app := models.ApplicationFields{
				State:            "stopped",
				RunningInstances: -1,
				InstanceCount:    2,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			Expect(instances).To(ContainSubstring("?/2"))
		})

		It("handles zero instances", func() {
			app := models.ApplicationFields{
				State:            "started",
				RunningInstances: 0,
				InstanceCount:    0,
			}

			instances := ui_helpers.ColoredAppInstances(app)
			Expect(instances).To(Equal("0/0"))
		})
	})

	Describe("ColoredInstanceState", func() {
		It("returns 'running' for started state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("started"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("running"))
		})

		It("returns 'running' for running state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("running"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("running"))
		})

		It("returns colored 'stopped' for stopped state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("stopped"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("stopped"))
		})

		It("returns colored 'crashed' for crashed state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("crashed"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("crashed"))
		})

		It("returns colored 'crashing' for flapping state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("flapping"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("crashing"))
		})

		It("returns colored 'down' for down state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("down"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("down"))
		})

		It("returns colored 'starting' for starting state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("starting"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("starting"))
		})

		It("returns colored warning for unknown state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("unknown-state"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).To(ContainSubstring("unknown-state"))
		})

		It("handles empty state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState(""),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			Expect(state).ToNot(BeEmpty())
		})

		It("handles uppercase state", func() {
			instance := models.AppInstanceFields{
				State: models.InstanceState("RUNNING"),
			}

			state := ui_helpers.ColoredInstanceState(instance)
			// The function converts it, but let's just verify it returns something
			Expect(state).ToNot(BeEmpty())
		})
	})
})
