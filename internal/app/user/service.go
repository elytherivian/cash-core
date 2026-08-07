package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"cash-core/internal/common"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(ctx context.Context, request CreateRequest) (*User, error)
	Register(ctx context.Context, req RegisterUserRequest) (*User, error)
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Create(ctx context.Context, request CreateRequest) (*User, error) {
	request.Normalize()
	if length := utf8.RuneCountInString(request.Username); length < 1 || length > 50 {
		return nil, fmt.Errorf("%w: username length must be between 1 and 50", common.ErrInvalidInput)
	}
	if len(request.Password) < 8 || len(request.Password) > 72 {
		return nil, fmt.Errorf("%w: password length must be between 8 and 72 bytes", common.ErrInvalidInput)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	value := &User{
		ID: uuid.New(), Username: request.Username, PasswordHash: string(passwordHash),
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, value); err != nil {
		if errors.Is(err, common.ErrConflict) {
			return nil, common.NewBusinessError(
				common.CodeRegisterUserAlreadyExists,
				"user already exists",
				err,
			)
		}
		return nil, err
	}
	return value, nil
}

func (s *service) Register(ctx context.Context, req RegisterUserRequest) (*User, error) {
	req.Normalize()
	if length := utf8.RuneCountInString(req.Username); length < 1 || length > 50 {
		return nil, fmt.Errorf("%w: username length must be between 1 and 50", common.ErrInvalidInput)
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		return nil, fmt.Errorf("%w: password length must be between 8 and 72 bytes", common.ErrInvalidInput)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &User{
		ID: uuid.New(), Username: req.Username, PasswordHash: string(passwordHash),
		Lifecycle: common.Lifecycle{IsActive: true},
	}
	if err := s.repository.Create(ctx, user); err != nil {
		if errors.Is(err, common.ErrConflict) {
			return nil, common.NewBusinessError(
				common.CodeRegisterUserAlreadyExists,
				"user already exists",
				err,
			)
		}
		return nil, err
	}
	return user, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id, time.Now().UTC())
}
