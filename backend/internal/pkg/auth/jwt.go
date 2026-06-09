package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// JWTClaims represents the JWT payload for authenticated users.
type JWTClaims struct {
	UserID   uint64 `json:"user_id"`
	TenantID uint64 `json:"tenant_id"`
	DeptID   uint64 `json:"dept_id"`
	Username string `json:"username"`
	DataScope int16 `json:"data_scope"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a signed JWT access token bound to the
// given system audience ("platform" or "shop"). The audience is
// verified by JWTAuthMiddleware to prevent cross-system token reuse.
func GenerateAccessToken(claims JWTClaims, secret string, ttl time.Duration, issuer, audience string) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID:        generateJTI(),
		Issuer:    issuer,
		Subject:   strconv.FormatUint(claims.UserID, 10),
		Audience:  jwt.ClaimStrings{audience},
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a signed JWT refresh token bound to
// the given system audience, suffixed with "-refresh".
func GenerateRefreshToken(claims JWTClaims, secret string, ttl time.Duration, issuer, audience string) (string, error) {
	now := time.Now()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ID:        generateJTI(),
		Issuer:    issuer,
		Subject:   strconv.FormatUint(claims.UserID, 10),
		Audience:  jwt.ClaimStrings{audience + "-refresh"},
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken parses and validates a JWT token string, returning the claims.
func ParseToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// ExtractAudience returns the first matching audience from the claims,
// stripping a "-refresh" suffix. Returns "" if no match.
func ExtractAudience(claims *JWTClaims, allowed ...string) string {
	if claims == nil || len(claims.Audience) == 0 {
		return ""
	}
	want := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		want[a] = struct{}{}
	}
	for _, raw := range claims.Audience {
		base := raw
		if i := len(base) - len("-refresh"); i > 0 && base[i:] == "-refresh" {
			base = base[:i]
		}
		if _, ok := want[base]; ok {
			return base
		}
	}
	return ""
}

// BlacklistToken adds a token to the Redis blacklist with remaining TTL.
func BlacklistToken(ctx context.Context, rdb redis.Cmdable, jti string, expiresAt time.Time) error {
	if expiresAt.IsZero() || time.Until(expiresAt) <= 0 {
		return nil
	}
	key := fmt.Sprintf("token:blacklist:%s", jti)
	return rdb.Set(ctx, key, "1", time.Until(expiresAt)).Err()
}

// IsTokenBlacklisted checks whether a token has been blacklisted.
func IsTokenBlacklisted(ctx context.Context, rdb redis.Cmdable, jti string) (bool, error) {
	key := fmt.Sprintf("token:blacklist:%s", jti)
	val, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// generateJTI generates a cryptographically random token ID.
func generateJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
