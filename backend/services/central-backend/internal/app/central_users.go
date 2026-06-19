package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"mercadia.dev/pos/services/central-backend/internal/domain"
)

var (
	ErrInvalidCentralUserCommand = errors.New("invalid central user command")
	ErrCentralUserConflict       = errors.New("central user conflict")
)

type CentralUsersService struct {
	users CentralUserRepository
	now   func() time.Time
}

func NewCentralUsersService(users CentralUserRepository) *CentralUsersService {
	return &CentralUsersService{
		users: users,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

type CreateCentralUserCommand struct {
	UserID      string
	Email       string
	DisplayName string
	Password    string
	Roles       []domain.CentralRole
	Session     SessionResult
}

type UpdateCentralUserCommand struct {
	UserID      string
	DisplayName *string
	Password    *string
	Roles       []domain.CentralRole
	Active      *bool
	Session     SessionResult
}

func (s *CentralUsersService) ListUsers(ctx context.Context, session SessionResult) ([]domain.CentralUser, error) {
	if err := CheckCentralPermission(session.Roles, PermissionUsersManage); err != nil {
		return nil, err
	}
	return s.users.ListUsers(ctx)
}

func (s *CentralUsersService) GetUser(ctx context.Context, userID string, session SessionResult) (domain.CentralUser, error) {
	if err := CheckCentralPermission(session.Roles, PermissionUsersManage); err != nil {
		return domain.CentralUser{}, err
	}
	return s.users.FindUser(ctx, userID)
}

func (s *CentralUsersService) CreateUser(ctx context.Context, command CreateCentralUserCommand) (domain.CentralUser, error) {
	if err := CheckCentralPermission(command.Session.Roles, PermissionUsersManage); err != nil {
		return domain.CentralUser{}, err
	}
	if command.UserID == "" || command.Email == "" || command.Password == "" || len(command.Roles) == 0 {
		return domain.CentralUser{}, ErrInvalidCentralUserCommand
	}

	if _, err := s.users.FindUserByEmail(ctx, command.Email); err == nil {
		return domain.CentralUser{}, ErrCentralUserConflict
	} else if !errors.Is(err, ErrCentralUserNotFound) {
		return domain.CentralUser{}, err
	}

	passwordHash, err := HashPassword(command.Password)
	if err != nil {
		return domain.CentralUser{}, err
	}

	user, err := domain.NewCentralUser(domain.CentralUser{
		ID:           command.UserID,
		Email:        command.Email,
		DisplayName:  command.DisplayName,
		PasswordHash: passwordHash,
		Roles:        command.Roles,
		CreatedAt:    s.now(),
	})
	if err != nil {
		return domain.CentralUser{}, err
	}
	if err := s.users.SaveUser(ctx, user); err != nil {
		return domain.CentralUser{}, err
	}
	return user, nil
}

func (s *CentralUsersService) UpdateUser(ctx context.Context, command UpdateCentralUserCommand) (domain.CentralUser, error) {
	if err := CheckCentralPermission(command.Session.Roles, PermissionUsersManage); err != nil {
		return domain.CentralUser{}, err
	}
	if command.UserID == "" {
		return domain.CentralUser{}, ErrInvalidCentralUserCommand
	}

	user, err := s.users.FindUser(ctx, command.UserID)
	if err != nil {
		return domain.CentralUser{}, err
	}

	if command.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*command.DisplayName)
		if user.DisplayName == "" {
			return domain.CentralUser{}, ErrInvalidCentralUserCommand
		}
	}
	if command.Password != nil {
		if *command.Password == "" {
			return domain.CentralUser{}, ErrInvalidCentralUserCommand
		}
		passwordHash, err := HashPassword(*command.Password)
		if err != nil {
			return domain.CentralUser{}, err
		}
		user.PasswordHash = passwordHash
	}
	if command.Roles != nil {
		if len(command.Roles) == 0 {
			return domain.CentralUser{}, ErrInvalidCentralUserCommand
		}
		user.Roles = append([]domain.CentralRole(nil), command.Roles...)
	}
	if command.Active != nil {
		user.Active = *command.Active
	}

	if err := s.users.SaveUser(ctx, user); err != nil {
		return domain.CentralUser{}, err
	}
	return user, nil
}
