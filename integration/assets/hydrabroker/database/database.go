package database

import (
	"fmt"
	"sync"

	uuid "github.com/google/uuid"
)

type ID string

type Database struct {
	data  map[ID]interface{}
	mutex sync.Mutex
}

func NewID() ID {
	rawGUID := uuid.New()

	return ID(rawGUID.String())
}

func NewDatabase() *Database {
	return &Database{data: make(map[ID]interface{})}
}

func (db *Database) Create(id ID, value interface{}) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.data[id]; ok {
		panic(fmt.Sprintf("duplicate ID: %s", id))
	}
	db.data[id] = value
}

func (db *Database) Retrieve(id ID) (interface{}, bool) {
	db.mutex.Lock() // Tests do not fail if this lock is removed
	defer db.mutex.Unlock()

	value, ok := db.data[id]
	return value, ok
}

func (db *Database) Update(id ID, value interface{}) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, ok := db.data[id]; !ok {
		panic(fmt.Sprintf("entry not found: %s", id))
	}
	db.data[id] = value
}

func (db *Database) Delete(id ID) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	delete(db.data, id)
}

func (db *Database) List() []ID {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	var ids []ID
	for k := range db.data {
		ids = append(ids, k)
	}

	return ids
}
