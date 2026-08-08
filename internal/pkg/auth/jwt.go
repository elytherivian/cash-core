package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cash-core/internal/common"
	"cash-core/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type TokenPair struct {
	UserID           uuid.UUID `json:"user_id"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	ExpiresIn        int64     `json:"expires_in"`
	RefreshExpiresIn int64     `json:"refresh_expires_in"`
}

type UserStateStore interface {
	AuthenticationState(ctx context.Context, userID uuid.UUID) (version int64, active bool, err error)
}

type Manager struct {
	issuer          string
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	userStates      UserStateStore
	now             func() time.Time
}

func NewManager(cfg config.Auth, userStates UserStateStore) *Manager {
	return &Manager{
		issuer: cfg.Issuer, secret: []byte(cfg.JWTSecret),
		accessTokenTTL: cfg.AccessTokenTTL, refreshTokenTTL: cfg.RefreshTokenTTL,
		userStates: userStates, now: time.Now,
	}
}

func (m *Manager) Issue(userID uuid.UUID, version int64) (TokenPair, error) {
	now := m.now().UTC()
	accessToken, err := m.sign(userID, version, accessTokenType, now, m.accessTokenTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign access token: %w", err)
	}
	refreshToken, err := m.sign(userID, version, refreshTokenType, now, m.refreshTokenTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("sign refresh token: %w", err)
	}
	return TokenPair{
		UserID: userID, AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer",
		ExpiresIn:        int64(m.accessTokenTTL.Seconds()),
		RefreshExpiresIn: int64(m.refreshTokenTTL.Seconds()),
	}, nil
}

// Verify 验证 access token，并确认用户当前仍有效且 token 版本未失效。
func (m *Manager) Verify(ctx context.Context, rawToken string) (uuid.UUID, error) {
	userID, _, err := m.verify(ctx, rawToken, accessTokenType)
	return userID, err
}

// Refresh 验证 refresh token 后签发一组新 token。
func (m *Manager) Refresh(ctx context.Context, rawToken string) (TokenPair, error) {
	userID, version, err := m.verify(ctx, rawToken, refreshTokenType)
	if err != nil {
		return TokenPair{}, err
	}
	return m.Issue(userID, version)
}

type claims struct {
	TokenType string `json:"token_type"`
	Version   int64  `json:"ver"`
	jwt.RegisteredClaims
}

func (m *Manager) sign(userID uuid.UUID, version int64, tokenType string, now time.Time, ttl time.Duration) (string, error) {
	value := claims{
		TokenType: tokenType,
		Version:   version,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: m.issuer, Subject: userID.String(), ID: uuid.NewString(),
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, value).SignedString(m.secret)
}

func (m *Manager) verify(ctx context.Context, rawToken, expectedType string) (uuid.UUID, int64, error) {
	value := new(claims)
	token, err := jwt.ParseWithClaims(
		strings.TrimSpace(rawToken),
		value,
		func(*jwt.Token) (any, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
		jwt.WithLeeway(5*time.Second),
	)
	if err != nil || !token.Valid || value.TokenType != expectedType {
		return uuid.Nil, 0, invalidTokenError()
	}
	userID, err := uuid.Parse(value.Subject)
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, 0, invalidTokenError()
	}
	version, active, err := m.userStates.AuthenticationState(ctx, userID)
	if err != nil || !active || version != value.Version {
		return uuid.Nil, 0, invalidTokenError()
	}
	return userID, version, nil
}

func invalidTokenError() error {
	return fmt.Errorf("%w: invalid or expired token", common.ErrUnauthenticated)
}
