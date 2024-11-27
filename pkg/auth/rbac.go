package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type Role string

const (
	RoleViewer   Role = "VIEWER"
	RoleOperator Role = "OPERATOR"
	RoleAdmin    Role = "ADMIN"
)

type Claims struct {
	Subject  string `json:"sub"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
}

type TokenValidator interface {
	ValidateToken(token string) (*Claims, error)
}

type MemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*Claims
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{
		tokens: make(map[string]*Claims),
	}
}

func (s *MemoryTokenStore) AddToken(token string, claims *Claims) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = claims
}

func (s *MemoryTokenStore) ValidateToken(token string) (*Claims, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claims, exists := s.tokens[token]
	if !exists {
		return nil, fmt.Errorf("%w: invalid token", core.ErrUnauthorized)
	}
	return claims, nil
}

type contextKey string

const claimsKey contextKey = "kestrel.claims"

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*Claims)
	return claims, ok
}

func HasRole(current, required Role) bool {
	switch required {
	case RoleViewer:
		return current == RoleViewer || current == RoleOperator || current == RoleAdmin
	case RoleOperator:
		return current == RoleOperator || current == RoleAdmin
	case RoleAdmin:
		return current == RoleAdmin
	default:
		return false
	}
}

func Middleware(validator TokenValidator, requiredRole Role, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization scheme", http.StatusUnauthorized)
			return
		}

		claims, err := validator.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		if !HasRole(claims.Role, requiredRole) {
			http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
			return
		}

		ctx := WithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
