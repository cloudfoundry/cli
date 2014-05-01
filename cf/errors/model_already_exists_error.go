package errors

import "fmt"

type ModelAlreadyExistsError struct {
	ModelType string
	ModelName string
}

func NewModelAlreadyExistsError(modelType, name string) *ModelAlreadyExistsError {
	return &ModelAlreadyExistsError{
		ModelType: modelType,
		ModelName: name,
	}
}

func (err *ModelAlreadyExistsError) Error() string {
	return fmt.Sprintf("%s %s already exists", err.ModelType, err.ModelName)
}
