package store_test

import (
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
)

var _ = Describe("Brokers", func() {
	var st *store.Store

	BeforeEach(func() {
		st = store.New()
	})

	It("can CRUD a broker", func() {
		cfg := config.BrokerConfiguration{
			Services: []config.Service{{
				Name: "baz",
				ID:   "1234",
			}},
			Username: "foo",
			Password: "bar",
		}

		By("creating")
		id := st.CreateBroker(cfg)

		By("retrieving")
		newCfg, ok := st.RetrieveBroker(id)
		Expect(ok).To(BeTrue())
		Expect(newCfg).To(Equal(cfg))

		By("updating")
		newCfg.Password = "lala"
		st.UpdateBroker(id, newCfg)
		updatedCfg, _ := st.RetrieveBroker(id)
		Expect(updatedCfg).To(Equal(newCfg))

		By("deleting")
		st.DeleteBroker(id)
		_, ok = st.RetrieveBroker(id)
		Expect(ok).To(BeFalse())
	})

	It("can list brokers", func() {
		var ids []store.BrokerID
		for i := 0; i < 10; i++ {
			ids = append(ids, st.CreateBroker(config.BrokerConfiguration{}))
		}

		Expect(st.ListBrokers()).To(ConsistOf(ids))
	})
})
