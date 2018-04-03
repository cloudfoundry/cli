package templatescompiler

import (
	"encoding/json"

	bireljob "github.com/cloudfoundry/bosh-cli/release/job"
	bierbrenderer "github.com/cloudfoundry/bosh-cli/templatescompiler/erbrenderer"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

type jobEvaluationContext struct {
	releaseJob           bireljob.Job
	releaseJobProperties *biproperty.Map
	jobProperties        biproperty.Map
	globalProperties     biproperty.Map
	deploymentName       string
	address              string
	uuidGen              boshuuid.Generator
	logger               boshlog.Logger
	logTag               string
}

// RootContext is exposed as an open struct in ERB templates.
// It must stay same to provide backwards compatible API.
type RootContext struct {
	Index      int        `json:"index"`
	ID         string     `json:"id"`
	AZ         string     `json:"az"`
	Bootstrap  bool       `json:"bootstrap"`
	JobContext jobContext `json:"job"`
	Deployment string     `json:"deployment"`
	Address    string     `json:"address,omitempty"`

	// Usually is accessed with <%= spec.networks.default.ip %>
	NetworkContexts map[string]networkContext `json:"networks"`

	//TODO: this should be a map[string]interface{}
	GlobalProperties  biproperty.Map  `json:"global_properties"`  // values from manifest's top-level properties
	ClusterProperties biproperty.Map  `json:"cluster_properties"` // values from instance group (deployment job) properties
	JobProperties     *biproperty.Map `json:"job_properties"`     // values from release job (aka template) properties
	DefaultProperties biproperty.Map  `json:"default_properties"` // values from release's job's spec
}

type jobContext struct {
	Name string `json:"name"`
}

type networkContext struct {
	IP      string `json:"ip"`
	Netmask string `json:"netmask"`
	Gateway string `json:"gateway"`
}

func NewJobEvaluationContext(
	releaseJob bireljob.Job,
	releaseJobProperties *biproperty.Map,
	jobProperties biproperty.Map,
	globalProperties biproperty.Map,
	deploymentName string,
	address string,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) bierbrenderer.TemplateEvaluationContext {
	return jobEvaluationContext{
		releaseJob:           releaseJob,
		releaseJobProperties: releaseJobProperties,
		jobProperties:        jobProperties,
		globalProperties:     globalProperties,
		deploymentName:       deploymentName,
		address:              address,
		uuidGen:              uuidGen,
		logTag:               "jobEvaluationContext",
		logger:               logger,
	}
}

func (ec jobEvaluationContext) MarshalJSON() ([]byte, error) {
	defaultProperties := ec.propertyDefaults(ec.releaseJob.Properties)
	var err error

	context := RootContext{
		Index:             0,
		AZ:                "unknown",
		Bootstrap:         true,
		JobContext:        jobContext{Name: ec.releaseJob.Name()},
		Deployment:        ec.deploymentName,
		NetworkContexts:   ec.buildNetworkContexts(),
		GlobalProperties:  ec.globalProperties,
		ClusterProperties: ec.jobProperties,
		JobProperties:     ec.releaseJobProperties,
		DefaultProperties: defaultProperties,
	}

	if len(ec.address) > 0 {
		context.Address = ec.address
	}

	context.ID, err = ec.uuidGen.Generate()
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Setting job eval context's ID to UUID: %#v", context)
	}

	ec.logger.Debug(ec.logTag, "Marshalling context %#v", context)

	jsonBytes, err := json.Marshal(context)
	if err != nil {
		return []byte{}, bosherr.WrapErrorf(err, "Marshalling job eval context: %#v", context)
	}

	return jsonBytes, nil
}

func (ec jobEvaluationContext) propertyDefaults(properties map[string]bireljob.PropertyDefinition) biproperty.Map {
	result := biproperty.Map{}
	for propertyKey, property := range properties {
		result[propertyKey] = property.Default
	}
	return result
}

func (ec jobEvaluationContext) buildNetworkContexts() map[string]networkContext {
	// IP is being returned by agent
	return map[string]networkContext{
		"default": networkContext{
			IP: "",
		},
	}
}
