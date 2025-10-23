package store

import (
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/database"
)

type BrokerID database.ID
type InstanceID database.ID
type BindingID database.ID

type Store struct {
	db *database.Database
}

func New() *Store {
	return &Store{db: database.NewDatabase()}
}
