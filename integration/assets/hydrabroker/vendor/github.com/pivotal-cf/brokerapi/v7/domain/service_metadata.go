package domain

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
)

type ServiceMetadata struct {
	DisplayName         string `json:"displayName,omitempty"`
	ImageUrl            string `json:"imageUrl,omitempty"`
	LongDescription     string `json:"longDescription,omitempty"`
	ProviderDisplayName string `json:"providerDisplayName,omitempty"`
	DocumentationUrl    string `json:"documentationUrl,omitempty"`
	SupportUrl          string `json:"supportUrl,omitempty"`
	Shareable           *bool  `json:"shareable,omitempty"`
	AdditionalMetadata  map[string]interface{}
}

func (sm ServiceMetadata) MarshalJSON() ([]byte, error) {
	type Alias ServiceMetadata

	b, err := json.Marshal(Alias(sm))
	if err != nil {
		return []byte{}, errors.Wrap(err, "unmarshallable content in AdditionalMetadata")
	}

	var m map[string]interface{}
	json.Unmarshal(b, &m)
	delete(m, additionalMetadataName)

	for k, v := range sm.AdditionalMetadata {
		m[k] = v
	}
	return json.Marshal(m)
}

func (sm *ServiceMetadata) UnmarshalJSON(data []byte) error {
	type Alias ServiceMetadata

	if err := json.Unmarshal(data, (*Alias)(sm)); err != nil {
		return err
	}

	additionalMetadata := map[string]interface{}{}
	if err := json.Unmarshal(data, &additionalMetadata); err != nil {
		return err
	}

	for _, jsonName := range GetJsonNames(reflect.ValueOf(sm).Elem()) {
		if jsonName == additionalMetadataName {
			continue
		}
		delete(additionalMetadata, jsonName)
	}

	if len(additionalMetadata) > 0 {
		sm.AdditionalMetadata = additionalMetadata
	}
	return nil
}
