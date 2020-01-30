package ccv3_test

import (
	"encoding/json"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("quota limits", func() {
	Describe("AppLimit", func() {
		DescribeTable("MarshalJSON",
			func(appLimit AppLimit, expectedBytes []byte) {
				bytes, err := json.Marshal(appLimit)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(Equal(expectedBytes))
			},
			Entry("total memory", AppLimit{TotalMemory: &types.NullInt{IsSet: true, Value: 1}}, []byte(`{"total_memory_in_mb":1}`)),
			Entry("total memory", AppLimit{TotalMemory: nil}, []byte(`{}`)),
			Entry("total memory", AppLimit{TotalMemory: &types.NullInt{IsSet: false}}, []byte(`{"total_memory_in_mb":null}`)),
			Entry("instance memory", AppLimit{InstanceMemory: &types.NullInt{IsSet: true, Value: 1}}, []byte(`{"per_process_memory_in_mb":1}`)),
			Entry("instance memory", AppLimit{InstanceMemory: nil}, []byte(`{}`)),
			Entry("instance memory", AppLimit{InstanceMemory: &types.NullInt{IsSet: false}}, []byte(`{"per_process_memory_in_mb":null}`)),
			Entry("total app instances", AppLimit{TotalAppInstances: &types.NullInt{IsSet: true, Value: 1}}, []byte(`{"total_instances":1}`)),
			Entry("total app instances", AppLimit{TotalAppInstances: nil}, []byte(`{}`)),
			Entry("total app instances", AppLimit{TotalAppInstances: &types.NullInt{IsSet: false}}, []byte(`{"total_instances":null}`)),
		)

		DescribeTable("UnmarshalJSON",
			func(givenBytes []byte, expectedStruct AppLimit) {
				var actualStruct AppLimit
				err := json.Unmarshal(givenBytes, &actualStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualStruct).To(Equal(expectedStruct))
			},
			Entry(
				"no null values",
				[]byte(`{"total_memory_in_mb":1,"per_process_memory_in_mb":2,"total_instances":3}`),
				AppLimit{
					TotalMemory:       &types.NullInt{IsSet: true, Value: 1},
					InstanceMemory:    &types.NullInt{IsSet: true, Value: 2},
					TotalAppInstances: &types.NullInt{IsSet: true, Value: 3},
				}),
			Entry(
				"total memory is null",
				[]byte(`{"total_memory_in_mb":null,"per_process_memory_in_mb":2,"total_instances":3}`),
				AppLimit{
					TotalMemory:       &types.NullInt{IsSet: false, Value: 0},
					InstanceMemory:    &types.NullInt{IsSet: true, Value: 2},
					TotalAppInstances: &types.NullInt{IsSet: true, Value: 3},
				}),
			Entry(
				"per process memory is null",
				[]byte(`{"total_memory_in_mb":1,"per_process_memory_in_mb":null,"total_instances":3}`),
				AppLimit{
					TotalMemory:       &types.NullInt{IsSet: true, Value: 1},
					InstanceMemory:    &types.NullInt{IsSet: false, Value: 0},
					TotalAppInstances: &types.NullInt{IsSet: true, Value: 3},
				}),
			Entry(
				"total instances is null",
				[]byte(`{"total_memory_in_mb":1,"per_process_memory_in_mb":2,"total_instances":null}`),
				AppLimit{
					TotalMemory:       &types.NullInt{IsSet: true, Value: 1},
					InstanceMemory:    &types.NullInt{IsSet: true, Value: 2},
					TotalAppInstances: &types.NullInt{IsSet: false, Value: 0},
				}),
		)
	})

	Describe("ServiceLimit", func() {
		var (
			trueValue  = true
			falseValue = false
		)

		DescribeTable("MarshalJSON",
			func(serviceLimit ServiceLimit, expectedBytes []byte) {
				bytes, err := json.Marshal(serviceLimit)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(Equal(expectedBytes))
			},
			Entry("total service instances", ServiceLimit{TotalServiceInstances: &types.NullInt{IsSet: true, Value: 1}}, []byte(`{"total_service_instances":1}`)),
			Entry("total service instances", ServiceLimit{TotalServiceInstances: nil}, []byte(`{}`)),
			Entry("total service instances", ServiceLimit{TotalServiceInstances: &types.NullInt{IsSet: false}}, []byte(`{"total_service_instances":null}`)),
			Entry("paid service plans", ServiceLimit{PaidServicePlans: &trueValue}, []byte(`{"paid_services_allowed":true}`)),
			Entry("paid service plans", ServiceLimit{PaidServicePlans: nil}, []byte(`{}`)),
			Entry("paid service plans", ServiceLimit{PaidServicePlans: &falseValue}, []byte(`{"paid_services_allowed":false}`)),
		)

		DescribeTable("UnmarshalJSON",
			func(givenBytes []byte, expectedStruct ServiceLimit) {
				var actualStruct ServiceLimit
				err := json.Unmarshal(givenBytes, &actualStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualStruct).To(Equal(expectedStruct))
			},
			Entry(
				"no null values",
				[]byte(`{"total_service_instances":1,"paid_services_allowed":true}`),
				ServiceLimit{
					TotalServiceInstances: &types.NullInt{IsSet: true, Value: 1},
					PaidServicePlans:      &trueValue,
				}),
			Entry(
				"total service instances is null and paid services allowed is false",
				[]byte(`{"total_service_instances":null,"paid_services_allowed":false}`),
				ServiceLimit{
					TotalServiceInstances: &types.NullInt{IsSet: false, Value: 0},
					PaidServicePlans:      &falseValue,
				}),
		)
	})

	Describe("RouteLimit", func() {
		DescribeTable("MarshalJSON",
			func(routeLimit RouteLimit, expectedBytes []byte) {
				bytes, err := json.Marshal(routeLimit)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(Equal(expectedBytes))
			},
			Entry("total routes", RouteLimit{TotalRoutes: &types.NullInt{IsSet: true, Value: 1}}, []byte(`{"total_routes":1}`)),
			Entry("total routes", RouteLimit{TotalRoutes: nil}, []byte(`{}`)),
			Entry("total routes", RouteLimit{TotalRoutes: &types.NullInt{IsSet: false}}, []byte(`{"total_routes":null}`)),
			Entry("total reserved ports", RouteLimit{TotalReservedPorts: &types.NullInt{IsSet: true, Value: 1}}, []byte(`{"total_reserved_ports":1}`)),
			Entry("total reserved ports", RouteLimit{TotalReservedPorts: nil}, []byte(`{}`)),
			Entry("total reserved ports", RouteLimit{TotalReservedPorts: &types.NullInt{IsSet: false}}, []byte(`{"total_reserved_ports":null}`)),
		)

		DescribeTable("UnmarshalJSON",
			func(givenBytes []byte, expectedStruct RouteLimit) {
				var actualStruct RouteLimit
				err := json.Unmarshal(givenBytes, &actualStruct)
				Expect(err).ToNot(HaveOccurred())
				Expect(actualStruct).To(Equal(expectedStruct))
			},
			Entry(
				"no null values",
				[]byte(`{"total_routes":1,"total_reserved_ports":2}`),
				RouteLimit{
					TotalRoutes:        &types.NullInt{IsSet: true, Value: 1},
					TotalReservedPorts: &types.NullInt{IsSet: true, Value: 2},
				}),
			Entry(
				"total routes is null",
				[]byte(`{"total_routes":null,"total_reserved_ports":3}`),
				RouteLimit{
					TotalRoutes:        &types.NullInt{IsSet: false, Value: 0},
					TotalReservedPorts: &types.NullInt{IsSet: true, Value: 3},
				}),
			Entry(
				"total reserved ports is null",
				[]byte(`{"total_routes":4,"total_reserved_ports":null}`),
				RouteLimit{
					TotalRoutes:        &types.NullInt{IsSet: true, Value: 4},
					TotalReservedPorts: &types.NullInt{IsSet: false, Value: 0},
				}),
		)

	})
})
