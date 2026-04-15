package proxy

import (
	"crypto/ecdsa"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// ExtractJWT validates the Bearer JWT in the Authorization header, extracts
// claims into the request context, and populates a Principal.
// It returns (principal, flatClaims, error).
// On error the caller should respond 403 and not forward the request.
func ExtractJWT(r *http.Request, pubKey *ecdsa.PublicKey) (*Principal, map[string]any, error) {
	raw := bearerToken(r)
	if raw == "" {
		return nil, nil, fmt.Errorf("missing Authorization Bearer token")
	}

	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("invalid JWT: %w", err)
	}
	if !token.Valid {
		return nil, nil, fmt.Errorf("JWT is not valid")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("unexpected claims type")
	}

	// Build flat string map for route matching (ctxKeyJWTClaims).
	// This is a separate map used for route matching and policy evaluation;
	// Principal.Claims preserves the raw JWT values (arrays, nested types).
	flat := make(map[string]any, len(mapClaims))
	for k, v := range mapClaims {
		flat[k] = fmt.Sprintf("%v", v)
	}

	// Build raw claims map preserving original types for impersonation.
	rawClaims := make(map[string]any, len(mapClaims))
	maps.Copy(rawClaims, mapClaims)

	// Extract well-known identity claims
	principal := &Principal{
		Claims: rawClaims,
	}

	if sub, ok := mapClaims["sub"].(string); ok {
		principal.Subject = sub
		principal.User = sub
	}

	if iss, ok := mapClaims["iss"].(string); ok {
		principal.Issuer = iss
	}

	if aud, ok := mapClaims["aud"]; ok {
		principal.Audience = toStringSlice(aud)
	}

	if exp, ok := mapClaims["exp"].(float64); ok {
		principal.ExpiresAt = time.Unix(int64(exp), 0).UTC()
	}

	if groups, ok := mapClaims["groups"]; ok {
		principal.Groups = toStringSlice(groups)
	}

	if sas, ok := mapClaims["service_accounts"]; ok {
		principal.ServiceAccounts = toStringSlice(sas)
	}

	return principal, flat, nil
}

// bearerToken extracts the raw token string from an Authorization: Bearer header.
func bearerToken(r *http.Request) string {
	ah := r.Header.Get("Authorization")
	if len(ah) > 7 && strings.EqualFold(ah[:7], "bearer ") {
		return ah[7:]
	}
	return ""
}

// toStringSlice coerces an interface{} JWT claim value into []string.
func toStringSlice(v any) []string {
	switch val := v.(type) {
	case []any:
		out := make([]string, 0, len(val))
		for _, item := range val {
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	case string:
		return []string{val}
	default:
		return nil
	}
}

// jwtMiddleware wraps an http.Handler with JWT extraction + context injection.
// It does NOT enforce policy — policy enforcement happens in ServeHTTP.
// func jwtMiddleware(pubKey *ecdsa.PublicKey, next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		principal, flat, err := ExtractJWT(r, pubKey)
// 		if err != nil {
// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 			return
// 		}
// 		ctx := WithPrincipal(r.Context(), principal)
// 		ctx = WithJWTClaims(ctx, flat)
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }
