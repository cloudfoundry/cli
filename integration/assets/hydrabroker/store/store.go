package store

import (
	"sync"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"

	uuid "github.com/nu7hatch/gouuid"
)

type BrokerConfigurationStore struct {
	data  map[string]config.BrokerConfiguration
	mutex sync.Mutex
}

func New() *BrokerConfigurationStore {
	return &BrokerConfigurationStore{
		data: make(map[string]config.BrokerConfiguration),
	}
}

func (c *BrokerConfigurationStore) CreateBroker(cfg config.BrokerConfiguration) (string, error) {
	rawGUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	guid := rawGUID.String()

	c.mutex.Lock()
	c.data[guid] = cfg
	c.mutex.Unlock()

	return guid, nil
}

func (c *BrokerConfigurationStore) DeleteBroker(guid string) {
	c.mutex.Lock()
	delete(c.data, guid)
	c.mutex.Unlock()
}

func (c *BrokerConfigurationStore) GetBrokerConfiguration(guid string) (config.BrokerConfiguration, bool) {
	c.mutex.Lock()
	config, ok := c.data[guid]
	c.mutex.Unlock()

	return config, ok
}

func (c *BrokerConfigurationStore) ListBrokers() (result []string) {
	c.mutex.Lock()
	for k := range c.data {
		result = append(result, k)
	}
	c.mutex.Unlock()

	return
}
