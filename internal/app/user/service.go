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
	Register(ctx context.Context, req RegisterUserRequest) (*User, error)
	Delete(ctx context.Context, req DeleteUserRequest) error
	Restore(ctx context.Context, req RestoreUserRequest) error
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Register(ctx context.Context, req RegisterUserRequest) (*User, error) {
	req.Normalize()
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password, 8); err != nil {
		return nil, err
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

func (s *service) Delete(ctx context.Context, req DeleteUserRequest) error {
	req.Normalize()
	if err := validateUsername(req.Username); err != nil {
		return err
	}
	if err := validatePassword(req.Password, 1); err != nil {
		return err
	}

	user, err := s.repository.FindByUsername(ctx, req.Username)
	if err != nil {
		return err
	}
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		return err
	}
	return s.repository.Delete(ctx, user.ID, time.Now().UTC())
}

func (s *service) Restore(ctx context.Context, req RestoreUserRequest) error {
	req.Normalize()
	if err := validateUsername(req.Username); err != nil {
		return err
	}
	if err := validatePassword(req.Password, 1); err != nil {
		return err
	}

	user, err := s.repository.FindByUsernameIncludingDeleted(ctx, req.Username)
	if err != nil {
		return err
	}
	if user.IsActive || user.DeletedAt == nil {
		return fmt.Errorf("%w: user is not deleted", common.ErrConflict)
	}
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		return err
	}
	return s.repository.Restore(ctx, user.ID, time.Now().UTC())
}

func validateUsername(username string) error {
	if length := utf8.RuneCountInString(username); length < 1 || length > 50 {
		return fmt.Errorf("%w: username length must be between 1 and 50", common.ErrInvalidInput)
	}
	return nil
}

func validatePassword(password string, minimumLength int) error {
	if len(password) < minimumLength || len(password) > 72 {
		return fmt.Errorf(
			"%w: password length must be between %d and 72 bytes",
			common.ErrInvalidInput,
			minimumLength,
		)
	}
	return nil
}

func comparePassword(passwordHash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return fmt.Errorf("%w: invalid password", common.ErrUnauthenticated)
		}
		return fmt.Errorf("compare password hash: %w", err)
	}
	return nil
}
