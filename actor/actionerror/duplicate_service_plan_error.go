package actionerror

import "fmt"

type DuplicateServicePlanError struct {
	Name                string
	ServiceOfferingName string
	ServiceBrokerName   string
}

func (e DuplicateServicePlanError) Error() string {
	base := fmt.Sprintf("Service plan '%s' is provided by multiple service offerings", e.Name)
	requiredFlag := "Specify an offering by using the '-o' flag"
	if e.ServiceOfferingName != "" {
		base = fmt.Sprintf("%s. "+
			"Service offering '%s' is provided by multiple service brokers", base, e.ServiceOfferingName)
		requiredFlag = "Specify a broker name by using the '-b' flag"
	}
	return fmt.Sprintf("%s. %s.", base, requiredFlag)
}
