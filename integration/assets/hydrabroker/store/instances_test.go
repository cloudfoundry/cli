package store_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/resources"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
	uuid "github.com/nu7hatch/gouuid"
)

var _ = Describe("Instances", func() {
	var (
		st     *store.Store
		broker store.BrokerID
	)

	randomID := func() store.InstanceID {
		rawGUID, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}

		return store.InstanceID(rawGUID.String())
	}

	BeforeEach(func() {
		st = store.New()

		broker = st.CreateBroker(config.BrokerConfiguration{
			Services: []config.Service{
				{
					ID: "service-offering-1-id",
					Plans: []config.Plan{
						{
							ID: "service-plan-1-id",
						},
						{
							ID: "service-plan-2-id",
						},
					},
				},
				{
					ID: "service-offering-2-id",
					Plans: []config.Plan{
						{
							ID: "service-plan-3-id",
						},
					},
				},
			},
		})
	})

	It("can CRUD a service instance", func() {
		By("creating")
		instanceID := randomID()
		details := resources.ServiceInstanceDetails{
			ServiceID:  "service-offering-1-id",
			PlanID:     "service-plan-1-id",
			Parameters: map[string]interface{}{"foo": "bar"},
		}
		err := st.CreateInstance(broker, instanceID, details)
		Expect(err).NotTo(HaveOccurred())

		By("retrieving")
		newDetails, err := st.RetrieveInstance(broker, instanceID)
		Expect(err).NotTo(HaveOccurred())
		Expect(newDetails).To(Equal(details))

		By("updating")
		newDetails.Parameters = map[string]interface{}{"baz": "quz"}
		err = st.UpdateInstance(broker, instanceID, newDetails)
		Expect(err).NotTo(HaveOccurred())
		Expect(st.RetrieveInstance(broker, instanceID)).To(Equal(newDetails))

		By("deleting")
		st.DeleteInstance(broker, instanceID)
		_, err = st.RetrieveInstance(broker, instanceID)
		Expect(err).To(MatchError(fmt.Sprintf("service instance not found: %s", instanceID)))
	})

	Describe("creating", func() {
		It("fails when broker not found", func() {
			Expect(st.CreateInstance("no-such-broker", "id", resources.ServiceInstanceDetails{})).
				To(MatchError("broker not found: no-such-broker"))
		})

		It("fails if the service ID is not in the catalog", func() {
			err := st.CreateInstance(broker, randomID(), resources.ServiceInstanceDetails{
				ServiceID: "no-such-service-offering",
			})
			Expect(err).To(MatchError("service offering ID not found in catalog: no-such-service-offering"))
		})

		It("fails if the plan ID is not in the catalog", func() {
			err := st.CreateInstance(broker, randomID(), resources.ServiceInstanceDetails{
				ServiceID: "service-offering-2-id",
				PlanID:    "no-such-service-plan",
			})
			Expect(err).To(MatchError("service plan ID 'no-such-service-plan' not found for service offering 'service-offering-2-id'"))
		})
	})

	Describe("creating", func() {
		It("fails when broker not found", func() {
			Expect(st.UpdateInstance("no-such-broker", "id", resources.ServiceInstanceDetails{})).
				To(MatchError("broker not found: no-such-broker"))
		})

		It("fails when service instance not found", func() {
			Expect(st.UpdateInstance(broker, "no-such-instance", resources.ServiceInstanceDetails{})).
				To(MatchError("service instance not found: no-such-instance"))
		})

		It("fails if the service ID is not in the catalog", func() {
			id := randomID()
			st.CreateInstance(broker, id, resources.ServiceInstanceDetails{
				ServiceID: "service-offering-1-id",
				PlanID:    "service-plan-1-id",
			})
			err := st.UpdateInstance(broker, id, resources.ServiceInstanceDetails{
				ServiceID: "no-such-service-offering",
				PlanID:    "service-plan-1-id",
			})
			Expect(err).To(MatchError("service offering ID not found in catalog: no-such-service-offering"))
		})

		It("fails if the plan ID is not in the catalog", func() {
			id := randomID()
			st.CreateInstance(broker, id, resources.ServiceInstanceDetails{
				ServiceID: "service-offering-1-id",
				PlanID:    "service-plan-1-id",
			})
			err := st.UpdateInstance(broker, id, resources.ServiceInstanceDetails{
				ServiceID: "service-offering-2-id",
				PlanID:    "no-such-service-plan",
			})
			Expect(err).To(MatchError("service plan ID 'no-such-service-plan' not found for service offering 'service-offering-2-id'"))
		})
	})

	Describe("retrieving", func() {
		It("fails when broker not found", func() {
			_, err := st.RetrieveInstance("no-such-broker", "id")
			Expect(err).To(MatchError("broker not found: no-such-broker"))
		})

		It("fails when service instance not found", func() {
			_, err := st.RetrieveInstance(broker, "no-such-instance")
			Expect(err).To(MatchError("service instance not found: no-such-instance"))
		})
	})

	Describe("Deleting", func() {
		It("fails when broker not found", func() {
			Expect(st.DeleteInstance("no-such-broker", "id")).
				To(MatchError("broker not found: no-such-broker"))
		})
	})
})
