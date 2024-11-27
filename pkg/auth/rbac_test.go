package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRBACMiddleware(t *testing.T) {
	store := NewMemoryTokenStore()
	store.AddToken("viewer-token", &Claims{Subject: "user-1", TenantID: "t1", Role: RoleViewer})
	store.AddToken("operator-token", &Claims{Subject: "user-2", TenantID: "t1", Role: RoleOperator})
	store.AddToken("admin-token", &Claims{Subject: "user-3", TenantID: "t1", Role: RoleAdmin})

	targetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok {
			t.Error("expected claims in context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok: " + string(claims.Role)))
	})

	operatorEndpoint := Middleware(store, RoleOperator, targetHandler)

	// 1. Missing Authorization header -> 401
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	operatorEndpoint.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	// 2. Viewer accessing Operator endpoint -> 403 Forbidden
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer viewer-token")
	w = httptest.NewRecorder()
	operatorEndpoint.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden for viewer, got %d", w.Code)
	}

	// 3. Operator accessing Operator endpoint -> 200 OK
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer operator-token")
	w = httptest.NewRecorder()
	operatorEndpoint.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for operator, got %d", w.Code)
	}

	// 4. Admin accessing Operator endpoint -> 200 OK (role escalation)
	req = httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w = httptest.NewRecorder()
	operatorEndpoint.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for admin, got %d", w.Code)
	}
}
