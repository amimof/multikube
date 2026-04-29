package infra

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

type NewTokenManagerOption func(*TokenManager)

type TokenManager struct {
	Key          *ecdsa.PrivateKey
	VerifyKey    *ecdsa.PublicKey
	AllowedAud   []string
	DefaultTTL   time.Duration
	MaxTTL       time.Duration
	DefaultAud   []string
	Issuer       string
	SigningKeyID string
}

type IssueResponse struct {
	AccessToken *jwt.Token
	ExpiresAt   time.Time
	KeyID       string
	TokenType   string
}

func (s *TokenManager) Issue(ctx context.Context, req *tokenv1.Token) (string, error) {
	if req.GetConfig().GetSubject() == "" {
		return "", errors.New("subject is req.GetConfig().Getired")
	}
	if req.GetConfig().GetTtl() == nil && s.DefaultTTL <= 0 {
		return "", errors.New("default ttl must be configured")
	}
	if s.Key == nil {
		return "", errors.New("signing key is not configured")
	}

	var ttl time.Duration
	if req.GetConfig().GetTtl() != nil {
		ttl = req.GetConfig().GetTtl().AsDuration()
	}
	if ttl <= 0 {
		ttl = s.DefaultTTL
	}
	if ttl > s.MaxTTL {
		return "", fmt.Errorf("req.GetConfig().Getested ttl %s exceeds max ttl %s", ttl, s.MaxTTL)
	}

	now := time.Now().UTC()
	exp := now.Add(ttl)

	aud := req.GetConfig().GetAudience()
	if len(aud) == 0 {
		aud = s.DefaultAud
	}
	if len(aud) == 0 {
		aud = []string{"multikube-proxy"}
	}

	claims := jwt.MapClaims{
		"iss": s.Issuer,
		"sub": req.GetConfig().GetSubject(),
		"aud": aud,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": exp.Unix(),
		"jti": uuid.NewString(),
	}

	if req.GetConfig().GetUsername() != "" {
		claims["preferred_username"] = req.GetConfig().GetUsername()
	}
	if len(req.GetConfig().GetGroups()) > 0 {
		claims["groups"] = req.GetConfig().GetGroups()
	}
	if len(req.GetConfig().GetServiceAccounts()) > 0 {
		claims["service_accounts"] = req.GetConfig().GetServiceAccounts()
	}
	if len(req.GetConfig().GetScopes()) > 0 {
		claims["scope"] = req.GetConfig().GetScopes()
	}
	if len(req.GetConfig().GetClusters()) > 0 {
		claims["clusters"] = req.GetConfig().GetClusters()
	}

	for k, v := range req.GetConfig().GetExtraClaims() {
		switch k {
		case "iss", "sub", "aud", "iat", "nbf", "exp", "jti", "groups", "service_accounts":
			return "", fmt.Errorf("extra claim %q is reserved", k)
		default:
			claims[k] = v
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.SigningKeyID
	token.Header["typ"] = "JWT"

	signed, err := token.SignedString(s.Key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (s *TokenManager) Verify(ctx context.Context, accessToken string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&jwt.MapClaims{},
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodES256 {
				return nil, fmt.Errorf("unexpected token signing method")
			}

			if s.VerifyKey == nil {
				return nil, errors.New("verification key is not configured")
			}

			return s.VerifyKey, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	var claims jwt.MapClaims
	switch c := token.Claims.(type) {
	case jwt.MapClaims:
		claims = c
	case *jwt.MapClaims:
		claims = *c
	default:
		return nil, errors.New("invalid token claims")
	}
	if typ, ok := token.Header["typ"].(string); ok && typ != "JWT" {
		return nil, errors.New("invalid token type")
	}
	if !claims.VerifyIssuer(s.Issuer, true) {
		return nil, errors.New("invalid token issuer")
	}
	if sub, ok := claims["sub"].(string); !ok || sub == "" {
		return nil, errors.New("invalid token subject")
	}

	allowedAud := s.AllowedAud
	if len(allowedAud) == 0 {
		allowedAud = s.DefaultAud
	}
	if len(allowedAud) == 0 {
		return nil, errors.New("allowed audiences are not configured")
	}

	var audValid bool
	for _, aud := range allowedAud {
		if claims.VerifyAudience(aud, true) {
			audValid = true
			break
		}
	}
	if !audValid {
		return nil, errors.New("invalid token audience")
	}

	return token, nil
}

func (s *TokenManager) Revoke(ctx context.Context, token *tokenv1.Token) error {
	return nil
}

func NewTokenManager(key *ecdsa.PrivateKey, opts ...NewTokenManagerOption) (*TokenManager, error) {
	t := &TokenManager{
		VerifyKey:    &key.PublicKey,
		AllowedAud:   []string{"multikube"},
		Issuer:       "https://auth.multikube.io",
		Key:          key,
		DefaultTTL:   time.Hour * 24,
		MaxTTL:       time.Hour * 72,
		DefaultAud:   []string{"multikube"},
		SigningKeyID: "key-2026-04",
	}

	for _, opt := range opts {
		opt(t)
	}

	return t, nil
}
