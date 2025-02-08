package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type JWTClaims struct {
	Subject   string   `json:"sub"`
	Role      Role     `json:"role"`
	TenantID  string   `json:"tenant_id,omitempty"`
	IssuedAt  int64    `json:"iat"`
	ExpiresAt int64    `json:"exp"`
}

type JWTManager struct {
	secret []byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{secret: []byte(secret)}
}

func (m *JWTManager) Sign(claims JWTClaims) (string, error) {
	headerJSON, _ := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	hB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	cB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	unsignedToken := fmt.Sprintf("%s.%s", hB64, cB64)
	sig := m.computeHMAC([]byte(unsignedToken))
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)

	return fmt.Sprintf("%s.%s", unsignedToken, sigB64), nil
}

func (m *JWTManager) Verify(tokenStr string) (*JWTClaims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt token format")
	}

	unsignedToken := fmt.Sprintf("%s.%s", parts[0], parts[1])
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("invalid signature encoding")
	}

	expectedSig := m.computeHMAC([]byte(unsignedToken))
	if !hmac.Equal(sig, expectedSig) {
		return nil, errors.New("token signature verification failed")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("invalid claims encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, err
	}

	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}

	return &claims, nil
}

func (m *JWTManager) computeHMAC(data []byte) []byte {
	h := hmac.New(sha256.New, m.secret)
	h.Write(data)
	return h.Sum(nil)
}
