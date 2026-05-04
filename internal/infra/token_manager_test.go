package infra

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

func TestTokenManagerVerifyTokenPurpose(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ctx := context.Background()
	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	resp, err := mgr.Issue(ctx, &tokenv1.Token{Config: &tokenv1.TokenConfig{Subject: "alice"}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	accessToken := resp.AccessToken
	refreshToken := resp.RefreshToken

	if _, err := mgr.VerifyAccessToken(ctx, accessToken); err != nil {
		t.Fatalf("verify access token: %v", err)
	}
	if _, err := mgr.VerifyRefreshToken(ctx, refreshToken); err != nil {
		t.Fatalf("verify refresh token: %v", err)
	}
	if _, err := mgr.VerifyRefreshToken(ctx, accessToken); err == nil {
		t.Fatal("expected access token to fail refresh verification")
	}
	if _, err := mgr.VerifyAccessToken(ctx, refreshToken); err == nil {
		t.Fatal("expected refresh token to fail access verification")
	}
}

func TestTokenManagerIssueUsesSeparateTTLs(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	ctx := context.Background()
	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	resp, err := mgr.Issue(ctx, &tokenv1.Token{Config: &tokenv1.TokenConfig{Subject: "alice"}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	accessToken := resp.AccessToken
	refreshToken := resp.RefreshToken
	if accessToken == refreshToken {
		t.Fatal("expected access and refresh tokens to differ")
	}

	accessClaims := decodeClaims(t, accessToken)
	refreshClaims := decodeClaims(t, refreshToken)

	if got := accessClaims[tokenTypeClaim]; got != accessTokenType {
		t.Fatalf("access typ = %v, want %q", got, accessTokenType)
	}
	if got := refreshClaims[tokenTypeClaim]; got != refreshTokenType {
		t.Fatalf("refresh typ = %v, want %q", got, refreshTokenType)
	}

	accessExp := int64(accessClaims["exp"].(float64))
	accessIat := int64(accessClaims["iat"].(float64))
	refreshExp := int64(refreshClaims["exp"].(float64))
	refreshIat := int64(refreshClaims["iat"].(float64))

	if got := time.Unix(accessExp, 0).Sub(time.Unix(accessIat, 0)); got != mgr.AccessTTL {
		t.Fatalf("access ttl = %s, want %s", got, mgr.AccessTTL)
	}
	if got := time.Unix(refreshExp, 0).Sub(time.Unix(refreshIat, 0)); got != mgr.RefreshTTL {
		t.Fatalf("refresh ttl = %s, want %s", got, mgr.RefreshTTL)
	}
	if refreshExp <= accessExp {
		t.Fatal("expected refresh token to expire after access token")
	}
	if _, ok := refreshClaims["groups"]; ok {
		t.Fatal("refresh token should not include access-only claims")
	}
}

func TestTokenManagerIssueWithoutTTLUsesDefaultAccessTTL(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	resp, err := mgr.Issue(context.Background(), &tokenv1.Token{Config: &tokenv1.TokenConfig{Subject: "alice"}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if resp.ExpiresAt.IsZero() {
		t.Fatal("expected exp to be set for default access ttl")
	}

	claims := decodeClaims(t, resp.AccessToken)
	accessExp := int64(claims["exp"].(float64))
	accessIat := int64(claims["iat"].(float64))
	if got := time.Unix(accessExp, 0).Sub(time.Unix(accessIat, 0)); got != mgr.AccessTTL {
		t.Fatalf("access ttl = %s, want %s", got, mgr.AccessTTL)
	}
}

func TestTokenManagerIssueZeroTTLNeverExpiresAccessToken(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	ttl := uint64(0)
	resp, err := mgr.Issue(context.Background(), &tokenv1.Token{Config: &tokenv1.TokenConfig{
		Subject: "alice",
		Ttl:     &ttl,
	}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if !resp.ExpiresAt.IsZero() {
		t.Fatalf("expires at = %v, want zero", resp.ExpiresAt)
	}

	accessClaims := decodeClaims(t, resp.AccessToken)
	if _, ok := accessClaims["exp"]; ok {
		t.Fatal("expected access token exp claim to be omitted")
	}

	refreshClaims := decodeClaims(t, resp.RefreshToken)
	refreshExp := int64(refreshClaims["exp"].(float64))
	refreshIat := int64(refreshClaims["iat"].(float64))
	if got := time.Unix(refreshExp, 0).Sub(time.Unix(refreshIat, 0)); got != mgr.RefreshTTL {
		t.Fatalf("refresh ttl = %s, want %s", got, mgr.RefreshTTL)
	}
	if _, err := mgr.VerifyAccessToken(context.Background(), resp.AccessToken); err != nil {
		t.Fatalf("verify access token: %v", err)
	}
}

func TestTokenManagerVerifyRejectsInvalidTokens(t *testing.T) {
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
			"iss":          mgr.Issuer,
			"sub":          "alice",
			"aud":          []string{mgr.DefaultAud},
			"iat":          time.Now().Add(-time.Minute).Unix(),
			"nbf":          time.Now().Add(-time.Minute).Unix(),
			"exp":          time.Now().Add(time.Hour).Unix(),
			"jti":          "test-jti",
			tokenTypeClaim: accessTokenType,
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

	tests := []struct {
		name  string
		token string
	}{
		{name: "wrong issuer", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims["iss"] = "https://wrong.example" })},
		{name: "wrong audience", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims["aud"] = []string{"other"} })},
		{name: "missing subject", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { delete(claims, "sub") })},
		{name: "expired token", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims["exp"] = time.Now().Add(-time.Minute).Unix() })},
		{name: "wrong type header", token: issue(t, func(_ jwt.MapClaims, tok *jwt.Token) { tok.Header["typ"] = "OTHER" })},
		{name: "wrong purpose", token: issue(t, func(claims jwt.MapClaims, _ *jwt.Token) { claims[tokenTypeClaim] = refreshTokenType })},
		{name: "non es256 token", token: issueHS256Token(t, mgr)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := mgr.VerifyAccessToken(ctx, tt.token); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestTokenManagerUserFromContextReturnsRoles(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	resp, err := mgr.Issue(context.Background(), &tokenv1.Token{Config: &tokenv1.TokenConfig{
		Subject: "alice",
		Roles:   []string{"admin"},
	}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	accessToken := resp.AccessToken

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", accessToken))

	username, roles, err := mgr.UserFromContext(ctx)
	if err != nil {
		t.Fatalf("user from context: %v", err)
	}
	if username != "alice" {
		t.Fatalf("username = %q, want %q", username, "alice")
	}
	if len(roles) != 1 || roles[0] != "admin" {
		t.Fatalf("roles = %v, want [admin]", roles)
	}
}

func TestTokenManagerStreamInterceptorInjectsRolesIntoContext(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	mgr, err := NewTokenManager(key)
	if err != nil {
		t.Fatalf("new token manager: %v", err)
	}

	resp, err := mgr.Issue(context.Background(), &tokenv1.Token{Config: &tokenv1.TokenConfig{
		Subject: "alice",
		Roles:   []string{"admin"},
	}})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	accessToken := resp.AccessToken

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", accessToken))
	stream := &stubServerStream{ctx: ctx}

	called := false
	err = mgr.Stream()(nil, stream, &grpc.StreamServerInfo{FullMethod: "/user.v1.UserService/Get"}, func(_ interface{}, ss grpc.ServerStream) error {
		called = true

		username, _ := ss.Context().Value(ContextKeyUsername).(string)
		roles, _ := ss.Context().Value(ContextKeyRoles).([]string)

		if username != "alice" {
			t.Fatalf("username = %q, want %q", username, "alice")
		}
		if len(roles) != 1 || roles[0] != "admin" {
			t.Fatalf("roles = %v, want [admin]", roles)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("stream interceptor: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func decodeClaims(t *testing.T, rawToken string) map[string]any {
	t.Helper()

	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token format: %q", rawToken)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}

	return claims
}

func issueHS256Token(t *testing.T, mgr *TokenManager) string {
	t.Helper()

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":          mgr.Issuer,
		"sub":          "alice",
		"aud":          []string{mgr.DefaultAud},
		"iat":          time.Now().Add(-time.Minute).Unix(),
		"nbf":          time.Now().Add(-time.Minute).Unix(),
		"exp":          time.Now().Add(time.Hour).Unix(),
		"jti":          "test-jti",
		tokenTypeClaim: accessTokenType,
	})
	tok.Header["typ"] = "JWT"

	signed, err := tok.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign hs256 token: %v", err)
	}

	return signed
}

type stubServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *stubServerStream) Context() context.Context {
	return s.ctx
}
