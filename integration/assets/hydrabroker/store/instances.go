package store

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/database"
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/resources"
)

func (st *Store) CreateInstance(brokerID BrokerID, instanceID InstanceID, details resources.ServiceInstanceDetails) error {
	brokerCfg, ok := st.RetrieveBroker(brokerID)
	if !ok {
		return fmt.Errorf("broker not found: %s", brokerID)
	}

	serviceOffering, err := findServiceOffering(brokerCfg, details.ServiceID)
	if err != nil {
		return err
	}

	_, err = findServicePlan(serviceOffering, details.PlanID)
	if err != nil {
		return err
	}

	st.db.Create(database.ID(instanceID), details)
	return nil
}

func (st *Store) RetrieveInstance(brokerID BrokerID, instanceID InstanceID) (resources.ServiceInstanceDetails, error) {
	_, ok := st.RetrieveBroker(brokerID)
	if !ok {
		return resources.ServiceInstanceDetails{}, fmt.Errorf("broker not found: %s", brokerID)
	}

	details, ok := st.db.Retrieve(database.ID(instanceID))
	if !ok {
		return resources.ServiceInstanceDetails{}, fmt.Errorf("service instance not found: %s", instanceID)
	}

	return details.(resources.ServiceInstanceDetails), nil
}

func (st *Store) UpdateInstance(brokerID BrokerID, instanceID InstanceID, details resources.ServiceInstanceDetails) error {
	brokerCfg, ok := st.RetrieveBroker(brokerID)
	if !ok {
		return fmt.Errorf("broker not found: %s", brokerID)
	}

	_, ok = st.db.Retrieve(database.ID(instanceID))
	if !ok {
		return fmt.Errorf("service instance not found: %s", instanceID)
	}

	serviceOffering, err := findServiceOffering(brokerCfg, details.ServiceID)
	if err != nil {
		return err
	}

	_, err = findServicePlan(serviceOffering, details.PlanID)
	if err != nil {
		return err
	}

	st.db.Update(database.ID(instanceID), details)
	return nil
}

func (st *Store) DeleteInstance(brokerID BrokerID, instanceID InstanceID) error {
	_, ok := st.RetrieveBroker(brokerID)
	if !ok {
		return fmt.Errorf("broker not found: %s", brokerID)
	}

	st.db.Delete(database.ID(instanceID))
	return nil
}

func findServiceOffering(cfg config.BrokerConfiguration, serviceOfferingID string) (config.Service, error) {
	for _, s := range cfg.Services {
		if s.ID == serviceOfferingID {
			return s, nil
		}
	}

	return config.Service{}, fmt.Errorf("service offering ID not found in catalog: %s", serviceOfferingID)
}

func findServicePlan(svc config.Service, servicePlanID string) (config.Plan, error) {
	for _, p := range svc.Plans {
		if p.ID == servicePlanID {
			return p, nil
		}
	}

	return config.Plan{}, fmt.Errorf("service plan ID '%s' not found for service offering '%s'", servicePlanID, svc.ID)
}
