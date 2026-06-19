package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrInvalidCentralUserInput = errors.New("invalid central user input")

type CentralRole string

const (
	CentralRoleViewer CentralRole = "central_viewer"
	CentralRoleAdmin  CentralRole = "central_admin"
)

type CentralUser struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	Roles        []CentralRole
	Active       bool
	CreatedAt    time.Time
}

func NewCentralUser(user CentralUser) (CentralUser, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	if user.ID == "" || user.Email == "" {
		return CentralUser{}, ErrInvalidCentralUserInput
	}
	if len(user.Roles) == 0 {
		return CentralUser{}, ErrInvalidCentralUserInput
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Email
	}
	user.Active = true
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	return user, nil
}

func (u CentralUser) HasRole(role CentralRole) bool {
	for _, candidate := range u.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}
