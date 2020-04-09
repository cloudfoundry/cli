package domain

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type ServicePlanMetadata struct {
	DisplayName        string            `json:"displayName,omitempty"`
	Bullets            []string          `json:"bullets,omitempty"`
	Costs              []ServicePlanCost `json:"costs,omitempty"`
	AdditionalMetadata map[string]interface{}
}

type ServicePlanCost struct {
	Amount map[string]float64 `json:"amount"`
	Unit   string             `json:"unit"`
}

func (spm *ServicePlanMetadata) UnmarshalJSON(data []byte) error {
	type Alias ServicePlanMetadata

	if err := json.Unmarshal(data, (*Alias)(spm)); err != nil {
		return err
	}

	additionalMetadata := map[string]interface{}{}
	if err := json.Unmarshal(data, &additionalMetadata); err != nil {
		return err
	}

	s := reflect.ValueOf(spm).Elem()
	for _, jsonName := range GetJsonNames(s) {
		if jsonName == additionalMetadataName {
			continue
		}
		delete(additionalMetadata, jsonName)
	}

	if len(additionalMetadata) > 0 {
		spm.AdditionalMetadata = additionalMetadata
	}
	return nil
}

func (spm ServicePlanMetadata) MarshalJSON() ([]byte, error) {
	type Alias ServicePlanMetadata

	b, err := json.Marshal(Alias(spm))
	if err != nil {
		return []byte{}, errors.Wrap(err, "unmarshallable content in AdditionalMetadata")
	}

	var m map[string]interface{}
	json.Unmarshal(b, &m)
	delete(m, additionalMetadataName)

	for k, v := range spm.AdditionalMetadata {
		m[k] = v
	}

	return json.Marshal(m)
}

func GetJsonNames(s reflect.Value) (res []string) {
	valType := s.Type()
	for i := 0; i < s.NumField(); i++ {
		field := valType.Field(i)
		tag := field.Tag
		jsonVal := tag.Get("json")
		if jsonVal != "" {
			components := strings.Split(jsonVal, ",")
			jsonName := components[0]
			res = append(res, jsonName)
		} else {
			res = append(res, field.Name)
		}
	}
	return res
}
