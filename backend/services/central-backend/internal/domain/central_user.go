package domain

import (
	"errors"
	"time"
)

var ErrInvalidCentralUserInput = errors.New("invalid central user input")

type CentralUser struct {
	ID          string
	Email       string
	DisplayName string
	Active      bool
	CreatedAt   time.Time
}

func NewCentralUser(user CentralUser) (CentralUser, error) {
	if user.ID == "" || user.Email == "" {
		return CentralUser{}, ErrInvalidCentralUserInput
	}
	user.Active = true
	return user, nil
}
