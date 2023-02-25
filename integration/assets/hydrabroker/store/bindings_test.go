package store_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/resources"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
	uuid "github.com/nu7hatch/gouuid"
)

var _ = Describe("Bindings", func() {
	var (
		st       *store.Store
		broker   store.BrokerID
		instance store.InstanceID
	)

	randomID := func() store.BindingID {
		rawGUID, err := uuid.NewV4()
		if err != nil {
			panic(err)
		}

		return store.BindingID(rawGUID.String())
	}

	BeforeEach(func() {
		st = store.New()

		broker = st.CreateBroker(config.BrokerConfiguration{
			Services: []config.Service{{
				ID: "1234",
				Plans: []config.Plan{{
					ID: "5678",
				}},
			}},
		})

		instance = "fake-instance-id"
		st.CreateInstance(broker, instance, resources.ServiceInstanceDetails{
			ServiceID: "1234",
			PlanID:    "5678",
		})
	})

	It("can CRD a binding", func() {
		By("creating")
		bindingID := randomID()
		details := resources.BindingDetails{
			Parameters: map[string]interface{}{"foo": "bar"},
		}
		st.CreateBinding(broker, instance, bindingID, details)

		By("retrieving")
		retrieved, err := st.RetrieveBinding(broker, instance, bindingID)
		Expect(err).NotTo(HaveOccurred())
		Expect(retrieved).To(Equal(details))

		By("deleting")
		st.DeleteBinding(broker, instance, bindingID)
		_, err = st.RetrieveBinding(broker, instance, bindingID)
		Expect(err).To(MatchError(fmt.Sprintf("service binding not found: %s", bindingID)))
	})

	Describe("creating", func() {
		It("fails when broker not found", func() {
			Expect(st.CreateBinding("no-such-broker", "id", "id", resources.BindingDetails{})).
				To(MatchError("broker not found: no-such-broker"))
		})

		It("fails when instance not found", func() {
			Expect(st.CreateBinding(broker, "no-such-instance", "id", resources.BindingDetails{})).
				To(MatchError("service instance not found: no-such-instance"))
		})
	})

	Describe("retrieving", func() {
		It("fails when broker not found", func() {
			_, err := st.RetrieveBinding("no-such-broker", "id", "id")
			Expect(err).To(MatchError("broker not found: no-such-broker"))
		})

		It("fails when instance not found", func() {
			_, err := st.RetrieveBinding(broker, "no-such-instance", "id")
			Expect(err).To(MatchError("service instance not found: no-such-instance"))
		})
	})

	Describe("deleting", func() {
		It("fails when broker not found", func() {
			err := st.DeleteBinding("no-such-broker", "id", "id")
			Expect(err).To(MatchError("broker not found: no-such-broker"))
		})

		It("fails when instance not found", func() {
			err := st.DeleteBinding(broker, "no-such-instance", "id")
			Expect(err).To(MatchError("service instance not found: no-such-instance"))
		})

		It("does not fail when the binding is not found", func() {
			err := st.DeleteBinding(broker, instance, "no-such-binding")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
