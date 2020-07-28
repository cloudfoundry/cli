package store

import (
	"fmt"
	"sync"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/resources"
	uuid "github.com/nu7hatch/gouuid"
)

type brokerState struct {
	instances map[string]resources.ServiceInstanceDetails
}

type brokerStoreEntry struct {
	config config.BrokerConfiguration
	state  brokerState
}

type BrokerConfigurationStore struct {
	data  map[string]brokerStoreEntry
	mutex sync.Mutex
}

func New() *BrokerConfigurationStore {
	return &BrokerConfigurationStore{
		data: make(map[string]brokerStoreEntry),
	}
}

func (c *BrokerConfigurationStore) CreateBroker(cfg config.BrokerConfiguration) string {
	rawGUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	guid := rawGUID.String()

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[guid] = brokerStoreEntry{config: cfg}

	return guid
}

func (c *BrokerConfigurationStore) UpdateBroker(guid string, cfg config.BrokerConfiguration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[guid] = brokerStoreEntry{config: cfg}
}

func (c *BrokerConfigurationStore) DeleteBroker(guid string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, guid)
}

func (c *BrokerConfigurationStore) GetBrokerConfiguration(guid string) (config.BrokerConfiguration, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, ok := c.data[guid]

	return entry.config, ok
}

func (c *BrokerConfigurationStore) ListBrokers() (result []string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k := range c.data {
		result = append(result, k)
	}

	return
}

func (c *BrokerConfigurationStore) CreateServiceInstance(brokerGUID string, serviceInstanceGUID string, serviceInstance resources.ServiceInstanceDetails) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, ok := c.data[brokerGUID]
	if !ok {
		return fmt.Errorf("service broker not found: %s", brokerGUID)
	}

	if entry.state.instances == nil {
		entry.state.instances = make(map[string]resources.ServiceInstanceDetails)
	}

	entry.state.instances[serviceInstanceGUID] = serviceInstance
	c.data[brokerGUID] = entry

	return nil
}

func (c *BrokerConfigurationStore) RetrieveServiceInstance(brokerGUID string, serviceInstanceGUID string) (resources.ServiceInstanceDetails, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, ok := c.data[brokerGUID]
	if !ok {
		return resources.ServiceInstanceDetails{}, fmt.Errorf("service broker not found: %s", brokerGUID)
	}

	instance, ok := entry.state.instances[serviceInstanceGUID]
	if !ok {
		return resources.ServiceInstanceDetails{}, fmt.Errorf("service instance %s not found for broker %s", serviceInstanceGUID, brokerGUID)
	}

	return instance, nil
}
