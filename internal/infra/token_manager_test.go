package infra

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

func TestTokenManagerVerify(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ctx := context.Background()
	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	issue := func(t *testing.T, mutate func(claims jwt.MapClaims, token *jwt.Token)) string {
		t.Helper()

		claims := jwt.MapClaims{
			"iss": mgr.Issuer,
			"sub": "alice",
			"aud": []string{mgr.DefaultAud[0]},
			"iat": time.Now().Add(-time.Minute).Unix(),
			"nbf": time.Now().Add(-time.Minute).Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		}

		tok := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		tok.Header["typ"] = "JWT"
		if mutate != nil {
			mutate(claims, tok)
		}

		signed, err := tok.SignedString(key)
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}

		return signed
	}

	validToken, err := mgr.Issue(ctx, &tokenv1.Token{Config: &tokenv1.TokenConfig{Subject: "alice"}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		wantError bool
	}{
		{name: "valid issued token", token: validToken},
		{name: "wrong issuer", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims["iss"] = "https://wrong.example" }), wantError: true},
		{name: "wrong audience", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims["aud"] = []string{"other"} }), wantError: true},
		{name: "missing subject", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { delete(claims, "sub") }), wantError: true},
		{name: "expired token", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims["exp"] = time.Now().Add(-time.Minute).Unix() }), wantError: true},
		{name: "wrong type header", token: issue(t, func(_ jwt.MapClaims, tok *jwt.Token) { tok.Header["typ"] = "OTHER" }), wantError: true},
		{name: "non es256 token", token: issueHS256Token(t, mgr), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Verify(ctx, tt.token)
			if tt.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func issueHS256Token(t *testing.T, mgr *TokenManager) string {
	t.Helper()

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": mgr.Issuer,
		"sub": "alice",
		"aud": []string{mgr.DefaultAud[0]},
		"iat": time.Now().Add(-time.Minute).Unix(),
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tok.Header["typ"] = "JWT"

	signed, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign hs256 token: %v", err)
	}

	return signed
}
