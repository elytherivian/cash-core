package user

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"cash-core/internal/common"
	"cash-core/internal/pkg/auth"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(ctx context.Context, req RegisterUserRequest) (*User, error)
	Delete(ctx context.Context, req DeleteUserRequest) error
	Restore(ctx context.Context, req RestoreUserRequest) error
	Login(ctx context.Context, req LoginRequest) (auth.TokenPair, error)
	Refresh(ctx context.Context, req RefreshTokenRequest) (auth.TokenPair, error)
}

type service struct {
	repository Repository
	tokens     TokenService
}

type TokenService interface {
	Issue(userID uuid.UUID, version int64) (auth.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (auth.TokenPair, error)
}

func NewService(repository Repository, tokens TokenService) Service {
	return &service{repository: repository, tokens: tokens}
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

func (s *service) Login(ctx context.Context, req LoginRequest) (auth.TokenPair, error) {
	req.Normalize()
	if err := validateUsername(req.Username); err != nil {
		return auth.TokenPair{}, err
	}
	if err := validatePassword(req.Password, 1); err != nil {
		return auth.TokenPair{}, err
	}

	user, err := s.repository.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			return auth.TokenPair{}, invalidCredentialsError()
		}
		return auth.TokenPair{}, err
	}
	if err := comparePassword(user.PasswordHash, req.Password); err != nil {
		if errors.Is(err, common.ErrUnauthenticated) {
			return auth.TokenPair{}, invalidCredentialsError()
		}
		return auth.TokenPair{}, err
	}
	if s.tokens == nil {
		return auth.TokenPair{}, errors.New("token service is not configured")
	}
	return s.tokens.Issue(user.ID, user.UpdatedAt.UTC().UnixMicro())
}

func (s *service) Refresh(ctx context.Context, req RefreshTokenRequest) (auth.TokenPair, error) {
	req.Normalize()
	if req.RefreshToken == "" {
		return auth.TokenPair{}, fmt.Errorf("%w: refresh token is required", common.ErrInvalidInput)
	}
	if s.tokens == nil {
		return auth.TokenPair{}, errors.New("token service is not configured")
	}
	return s.tokens.Refresh(ctx, req.RefreshToken)
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

func invalidCredentialsError() error {
	return fmt.Errorf("%w: invalid username or password", common.ErrUnauthenticated)
}
