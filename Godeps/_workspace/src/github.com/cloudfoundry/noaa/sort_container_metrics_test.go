package noaa_test

import (
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SortContainerMetrics", func() {
	var messages []*events.ContainerMetric

	BeforeEach(func() {
		messages = []*events.ContainerMetric{
			&events.ContainerMetric{
				ApplicationId: proto.String("appId"),
				InstanceIndex: proto.Int32(2),
			},
			&events.ContainerMetric{
				ApplicationId: proto.String("appId"),
				InstanceIndex: proto.Int32(1),
			},
		}
	})

	It("sorts container metrics by instance index", func() {
		sortedMessages := noaa.SortContainerMetrics(messages)

		Expect(sortedMessages[0].GetInstanceIndex()).To(Equal(int32(1)))
		Expect(sortedMessages[1].GetInstanceIndex()).To(Equal(int32(2)))
	})
})
