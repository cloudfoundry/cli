package resources_test

import (
	. "cf/api/resources"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testtime "testhelpers/time"
)

var _ = Describe("Event resources", func() {
	It("unmarshals app crash events", func() {
		resource := new(EventResourceNewV2)
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
		resource := new(EventResourceNewV2)
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
		Expect(eventFields.Description).To(Equal("instances: 1, memory: 256, command: PRIVATE DATA HIDDEN, environment_json: PRIVATE DATA HIDDEN"))

		resource = new(EventResourceNewV2)
		err = json.Unmarshal([]byte(`
		{
		  "metadata": {
			"guid": "event-1-guid"
		  },
		  "entity": {
			"type": "audit.app.update",
			"timestamp": "2014-01-21T00:20:11+00:00",
			"metadata": {
			  "request": {
				"state": "STOPPED"
			  }
			}
		  }
		}`), &resource)

		Expect(err).NotTo(HaveOccurred())

		eventFields = resource.ToFields()
		Expect(eventFields.Guid).To(Equal("event-1-guid"))
		Expect(eventFields.Name).To(Equal("audit.app.update"))
		Expect(eventFields.Timestamp).To(Equal(testtime.MustParse(eventTimestampFormat, "2014-01-21T00:20:11+00:00")))
		Expect(eventFields.Description).To(Equal(`state: STOPPED`))
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

const eventTimestampFormat = "2006-01-02T15:04:05-07:00"
