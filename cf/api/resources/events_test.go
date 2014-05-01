package resources_test

import (
	"encoding/json"
	. "github.com/cloudfoundry/cli/cf/api/resources"
	testtime "github.com/cloudfoundry/cli/testhelpers/time"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event resources", func() {
	var resource EventResource

	Describe("New V2 resources", func() {
		BeforeEach(func() {
			resource = new(EventResourceNewV2)
		})

		It("unmarshals app crash events", func() {
			err := json.Unmarshal([]byte(`
			{
			  "metadata": {
				"guid":"event-1-guid"
			  },
			  "entity": {
				"timestamp": "2013-10-07T16:51:07+00:00",
				"type": "app.crash",
				"metadata": {
				  "instance": "50dd66d3f8874b35988d23a25d19bfa0",
				  "index": 3,
				  "exit_status": -1,
				  "exit_description": "unknown",
				  "reason": "CRASHED"
				}
			  }
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("app.crash"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2013-10-07T16:51:07+00:00")))
			Expect(eventFields.Description).To(Equal(`index: 3, reason: CRASHED, exit_description: unknown, exit_status: -1`))
		})

		It("unmarshals app update events", func() {
			err := json.Unmarshal([]byte(`
			{
			  "metadata": {
				"guid": "event-1-guid"
			  },
			  "entity": {
				"type": "audit.app.update",
				"timestamp": "2014-01-21T00:20:11+00:00",
				"metadata": {
				  "request": {
				  	"state": "STOPPED",
					"command": "PRIVATE DATA HIDDEN",
					"instances": 1,
					"memory": 256,
					"environment_json": "PRIVATE DATA HIDDEN"
				  }
				}
			  }
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("audit.app.update"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00")))
			Expect(eventFields.Description).To(Equal("instances: 1, memory: 256, state: STOPPED, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN"))
		})

		It("unmarshals app delete events", func() {
			resource := new(EventResourceNewV2)
			err := json.Unmarshal([]byte(`
			{
			  "metadata": {
				"guid": "event-2-guid"
			  },
			  "entity": {
				"type": "audit.app.delete-request",
				"timestamp": "2014-01-21T18:39:09+00:00",
				"metadata": {
				  "request": {
					"recursive": true
				  }
				}
			  }
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-2-guid"))
			Expect(eventFields.Name).To(Equal("audit.app.delete-request"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-21T18:39:09+00:00")))
			Expect(eventFields.Description).To(Equal("recursive: true"))
		})

		It("unmarshals the new v2 app create event", func() {
			resource := new(EventResourceNewV2)
			err := json.Unmarshal([]byte(`
			{
			  "metadata": {
				"guid": "event-1-guid"
			  },
			  "entity": {
				"type": "audit.app.create",
				"timestamp": "2014-01-22T19:34:16+00:00",
				"metadata": {
				  "request": {
					"name": "java-warz",
					"space_guid": "6cc20fec-0dee-4843-b875-b124bfee791a",
					"production": false,
					"environment_json": "PRIVATE DATA HIDDEN",
					"instances": 1,
					"disk_quota": 1024,
					"state": "STOPPED",
					"console": false
				  }
				}
			  }
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())

			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("audit.app.create"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-22T19:34:16+00:00")))
			Expect(eventFields.Description).To(Equal("disk_quota: 1024, instances: 1, state: STOPPED, environment_json: PRIVATE DATA HIDDEN"))
		})
	})

	Describe("Old V2 Resources", func() {
		BeforeEach(func() {
			resource = new(EventResourceOldV2)
		})

		It("unmarshals app crashed events", func() {
			err := json.Unmarshal([]byte(`
			{
			  "metadata": {
				"guid": "event-1-guid"
			  },
			  "entity": {
				"timestamp": "2014-01-22T19:34:16+00:00",
				"exit_status": 3,
				"instance_index": 4,
				"exit_description": "the exit description"
			  }
			}`), &resource)

			Expect(err).NotTo(HaveOccurred())
			eventFields := resource.ToFields()
			Expect(eventFields.Guid).To(Equal("event-1-guid"))
			Expect(eventFields.Name).To(Equal("app crashed"))
			Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-22T19:34:16+00:00")))
			Expect(eventFields.Description).To(Equal("instance: 4, reason: the exit description, exit_status: 3"))
		})
	})
})

const eventTimestampFormat = "2006-01-02T15:04:05-07:00"
