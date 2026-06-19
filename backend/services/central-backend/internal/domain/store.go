package domain

import (
	"errors"
	"time"
)

var ErrInvalidStoreInput = errors.New("invalid store input")

type Store struct {
	ID           string
	Name         string
	Region       string
	RegisteredAt time.Time
	UpdatedAt    time.Time
}

func NewStore(store Store) (Store, error) {
	if store.ID == "" || store.Name == "" {
		return Store{}, ErrInvalidStoreInput
	}
	if store.Region == "" {
		store.Region = "default"
	}
	return store, nil
}
