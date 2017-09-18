package helpers_test

import (
	. "code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppInstanceTable", func() {
	It("can parse app instance table from v3-app", func() {
		input := `
Showing health and status for app dora in org wut / space wut as admin...

name:              dora
requested state:   started
processes:         web:4/4
memory usage:      32M x 4
routes:            dora.bosh-lite.com
stack:             cflinuxfs2
buildpacks:        ruby 1.6.44

web:4/4
     state     since                    cpu    memory         disk
#0   running   2017-08-02 17:12:10 PM   0.0%   21.2M of 32M   84.5M of 1G
#1   running   2017-08-03 09:39:25 AM   0.2%   19.3M of 32M   84.5M of 1G
#2   running   2017-08-03 03:29:25 AM   0.1%   22.8M of 32M   84.5M of 1G
#3   running   2017-08-02 17:12:10 PM   0.2%   22.9M of 32M   84.5M of 1G

worker:1/1
     state     since                    cpu    memory      disk
#0   stopped   2017-08-02 17:12:10 PM   0.0%   0M of 32M   0M of 1G
`
		appInstanceTable := ParseV3AppProcessTable([]byte(input))
		Expect(appInstanceTable).To(Equal(AppTable{
			Processes: []AppProcessTable{
				{
					Title: "web:4/4",
					Instances: []AppInstanceRow{
						{Index: "#0", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.0%", Memory: "21.2M of 32M", Disk: "84.5M of 1G"},
						{Index: "#1", State: "running", Since: "2017-08-03 09:39:25 AM", CPU: "0.2%", Memory: "19.3M of 32M", Disk: "84.5M of 1G"},
						{Index: "#2", State: "running", Since: "2017-08-03 03:29:25 AM", CPU: "0.1%", Memory: "22.8M of 32M", Disk: "84.5M of 1G"},
						{Index: "#3", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.2%", Memory: "22.9M of 32M", Disk: "84.5M of 1G"},
					},
				},
				{
					Title: "worker:1/1",
					Instances: []AppInstanceRow{
						{Index: "#0", State: "stopped", Since: "2017-08-02 17:12:10 PM", CPU: "0.0%", Memory: "0M of 32M", Disk: "0M of 1G"},
					},
				},
			},
		}))
	})

	It("can parse app instance table from v3-scale", func() {
		input := `
Showing health and status for app dora in org wut / space wut as admin...

web:4/4
     state     since                    cpu    memory         disk
#0   running   2017-08-02 17:12:10 PM   0.0%   21.2M of 32M   84.5M of 1G
#1   running   2017-08-03 09:39:25 AM   0.2%   19.3M of 32M   84.5M of 1G
#2   running   2017-08-03 03:29:25 AM   0.1%   22.8M of 32M   84.5M of 1G
#3   running   2017-08-02 17:12:10 PM   0.2%   22.9M of 32M   84.5M of 1G
`
		appInstanceTable := ParseV3AppProcessTable([]byte(input))
		Expect(appInstanceTable).To(Equal(AppTable{
			Processes: []AppProcessTable{
				{
					Title: "web:4/4",
					Instances: []AppInstanceRow{
						{Index: "#0", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.0%", Memory: "21.2M of 32M", Disk: "84.5M of 1G"},
						{Index: "#1", State: "running", Since: "2017-08-03 09:39:25 AM", CPU: "0.2%", Memory: "19.3M of 32M", Disk: "84.5M of 1G"},
						{Index: "#2", State: "running", Since: "2017-08-03 03:29:25 AM", CPU: "0.1%", Memory: "22.8M of 32M", Disk: "84.5M of 1G"},
						{Index: "#3", State: "running", Since: "2017-08-02 17:12:10 PM", CPU: "0.2%", Memory: "22.9M of 32M", Disk: "84.5M of 1G"},
					},
				},
			},
		}))
	})
})
