package helpers_test

import (
	. "code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppInstanceTable", func() {
	It("can parse app instance table from app", func() {
		input := `
Showing health and status for app dora in org wut / space wut as admin...

name:              dora
requested state:   started
routes:            dora.bosh-lite.com
stack:             cflinuxfs2
buildpacks:        ruby 1.6.44

type:         web
instances:    4/4
memory usage: 32M
     state     since                    cpu    memory         disk          logging
#0   running   2017-08-02 17:12:10 PM   0.0%   21.2M of 32M   84.5M of 1G   5B/s of 1K/s
#1   running   2017-08-03 09:39:25 AM   0.2%   19.3M of 32M   84.5M of 1G   7B/s of 1K/s
#2   running   2017-08-03 03:29:25 AM   0.1%   22.8M of 32M   84.5M of 1G   10B/s of 1K/s
#3   running   2017-08-02 17:12:10 PM   0.2%   22.9M of 32M   84.5M of 1G   8B/s of 1K/s

type:         worker
instances:    1/1
memory usage: 32M
     state     since                    cpu    memory      disk       logging
#0   stopped   2017-08-02 17:12:10 PM   0.0%   0M of 32M   0M of 1G   0B/s of 1K/s
`
		appInstanceTable := ParseV3AppProcessTable([]byte(input))
		Expect(appInstanceTable).To(Equal(AppTable{
			Processes: []AppProcessTable{
				{
					Type:          "web",
					InstanceCount: "4/4",
					MemUsage:      "32M",
					Instances: []AppInstanceRow{
						{Index: "#0", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.0%", Memory: "21.2M of 32M", Disk: "84.5M of 1G", LogRate: "5B/s of 1K/s"},
						{Index: "#1", State: "running", Since: "2017-08-03 09:39:25 AM", CPU: "0.2%", Memory: "19.3M of 32M", Disk: "84.5M of 1G", LogRate: "7B/s of 1K/s"},
						{Index: "#2", State: "running", Since: "2017-08-03 03:29:25 AM", CPU: "0.1%", Memory: "22.8M of 32M", Disk: "84.5M of 1G", LogRate: "10B/s of 1K/s"},
						{Index: "#3", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.2%", Memory: "22.9M of 32M", Disk: "84.5M of 1G", LogRate: "8B/s of 1K/s"},
					},
				},
				{
					Type:          "worker",
					InstanceCount: "1/1",
					MemUsage:      "32M",
					Instances: []AppInstanceRow{
						{Index: "#0", State: "stopped", Since: "2017-08-02 17:12:10 PM", CPU: "0.0%", Memory: "0M of 32M", Disk: "0M of 1G", LogRate: "0B/s of 1K/s"},
					},
				},
			},
		}))
	})

	It("can parse app instance table from scale", func() {
		input := `
Showing health and status for app dora in org wut / space wut as admin...

type:         web
instances:    4/4
memory usage: 32M
     state     since                    cpu    memory         disk          logging
#0   running   2017-08-02 17:12:10 PM   0.0%   21.2M of 32M   84.5M of 1G   1.3K/s of 5K/s
#1   running   2017-08-03 09:39:25 AM   0.2%   19.3M of 32M   84.5M of 1G   1.2K/s of 5K/s
#2   running   2017-08-03 03:29:25 AM   0.1%   22.8M of 32M   84.5M of 1G   1.1K/s of 5K/s
#3   running   2017-08-02 17:12:10 PM   0.2%   22.9M of 32M   84.5M of 1G   1.2K/s of 5K/s
`
		appInstanceTable := ParseV3AppProcessTable([]byte(input))
		Expect(appInstanceTable).To(Equal(AppTable{
			Processes: []AppProcessTable{
				{
					Type:          "web",
					InstanceCount: "4/4",
					MemUsage:      "32M",
					Instances: []AppInstanceRow{
						{Index: "#0", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.0%", Memory: "21.2M of 32M", Disk: "84.5M of 1G", LogRate: "1.3K/s of 5K/s"},
						{Index: "#1", State: "running", Since: "2017-08-03 09:39:25 AM", CPU: "0.2%", Memory: "19.3M of 32M", Disk: "84.5M of 1G", LogRate: "1.2K/s of 5K/s"},
						{Index: "#2", State: "running", Since: "2017-08-03 03:29:25 AM", CPU: "0.1%", Memory: "22.8M of 32M", Disk: "84.5M of 1G", LogRate: "1.1K/s of 5K/s"},
						{Index: "#3", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.2%", Memory: "22.9M of 32M", Disk: "84.5M of 1G", LogRate: "1.2K/s of 5K/s"},
					},
				},
			},
		}))
	})
})
