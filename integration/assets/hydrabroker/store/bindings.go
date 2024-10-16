package store

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/database"
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/resources"
)

func (st *Store) CreateBinding(brokerID BrokerID, instanceID InstanceID, bindingID BindingID, details resources.BindingDetails) error {
	_, err := st.RetrieveInstance(brokerID, instanceID)
	if err != nil {
		return err
	}

	st.db.Create(database.ID(bindingID), details)
	return nil
}

func (st *Store) RetrieveBinding(brokerID BrokerID, instanceID InstanceID, bindingID BindingID) (resources.BindingDetails, error) {
	_, err := st.RetrieveInstance(brokerID, instanceID)
	if err != nil {
		return resources.BindingDetails{}, err
	}

	details, ok := st.db.Retrieve(database.ID(bindingID))
	if !ok {
		return resources.BindingDetails{}, fmt.Errorf("service binding not found: %s", bindingID)
	}

	return details.(resources.BindingDetails), nil
}

func (st *Store) DeleteBinding(brokerID BrokerID, instanceID InstanceID, bindingID BindingID) error {
	_, err := st.RetrieveInstance(brokerID, instanceID)
	if err != nil {
		return err
	}

	st.db.Delete(database.ID(bindingID))
	return nil
}
