package middleware

import (
	"context"
	"strings"

	"cash-core/internal/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDKey = "authenticated_user_id"

// TokenVerifier lets the authentication middleware stay independent from JWT,
// opaque tokens, sessions, or an external identity provider.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (uuid.UUID, error)
}

func Authentication(verifier TokenVerifier, responder common.Responder) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(strings.TrimSpace(c.GetHeader("Authorization")), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			responder.Error(c, common.ErrUnauthenticated)
			return
		}
		userID, err := verifier.Verify(c.Request.Context(), strings.TrimSpace(token))
		if err != nil || userID == uuid.Nil {
			responder.Error(c, common.ErrUnauthenticated)
			return
		}
		c.Set(UserIDKey, userID)
		c.Next()
	}
}

func AuthenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok
}
