package store

import (
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/database"
)

func (st *Store) CreateBroker(cfg config.BrokerConfiguration) BrokerID {
	id := database.NewID()
	st.db.Create(id, cfg)
	return BrokerID(id)
}

func (st *Store) RetrieveBroker(id BrokerID) (config.BrokerConfiguration, bool) {
	if cfg, ok := st.db.Retrieve(database.ID(id)); ok {
		return cfg.(config.BrokerConfiguration), true
	}
	return config.BrokerConfiguration{}, false
}

func (st *Store) UpdateBroker(id BrokerID, cfg config.BrokerConfiguration) {
	st.db.Update(database.ID(id), cfg)
}

func (st *Store) DeleteBroker(id BrokerID) {
	st.db.Delete(database.ID(id))
}

func (st *Store) ListBrokers() []BrokerID {
	var brokers []BrokerID
	for _, id := range st.db.List() {
		if cfg, ok := st.db.Retrieve(id); ok {
			switch cfg.(type) {
			case config.BrokerConfiguration:
				brokers = append(brokers, BrokerID(id))
			default:
			}
		}
	}

	return brokers
}
