package noaa_test

import (
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SortRecent", func() {
	var messages []*events.LogMessage

	BeforeEach(func() {
		messages = []*events.LogMessage{createLogMessage("hello", 2), createLogMessage("konnichiha", 1)}
	})

	It("sorts messages", func() {
		sortedMessages := noaa.SortRecent(messages)

		Expect(*sortedMessages[0].Timestamp).To(Equal(int64(1)))
		Expect(*sortedMessages[1].Timestamp).To(Equal(int64(2)))
	})

	It("sorts using a stable algorithm", func() {
		messages = append(messages, createLogMessage("guten tag", 1))

		sortedMessages := noaa.SortRecent(messages)

		Expect(sortedMessages[0].GetMessage()).To(Equal([]byte("konnichiha")))
		Expect(sortedMessages[1].GetMessage()).To(Equal([]byte("guten tag")))
		Expect(sortedMessages[2].GetMessage()).To(Equal([]byte("hello")))
	})
})
