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
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	MaxTTL       time.Duration
	DefaultAud   []string
	Issuer       string
	SigningKeyID string
}

const (
	tokenTypeClaim   = "typ"
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type IssueResponse struct {
	AccessToken *jwt.Token
	ExpiresAt   time.Time
	KeyID       string
	TokenType   string
}

func (s *TokenManager) Issue(ctx context.Context, req *tokenv1.Token) (string, string, error) {
	if req.GetConfig().GetSubject() == "" {
		return "", "", errors.New("subject is req.GetConfig().Getired")
	}
	if s.AccessTTL <= 0 {
		return "", "", errors.New("access ttl must be configured")
	}
	if s.RefreshTTL <= 0 {
		return "", "", errors.New("refresh ttl must be configured")
	}
	if s.Key == nil {
		return "", "", errors.New("signing key is not configured")
	}

	accessTTL := s.AccessTTL
	if req.GetConfig().GetTtl() != nil {
		accessTTL = req.GetConfig().GetTtl().AsDuration()
	}
	if accessTTL <= 0 {
		accessTTL = s.AccessTTL
	}
	if accessTTL > s.MaxTTL {
		return "", "", fmt.Errorf("req.GetConfig().Getested ttl %s exceeds max ttl %s", accessTTL, s.MaxTTL)
	}
	if s.RefreshTTL > s.MaxTTL {
		return "", "", fmt.Errorf("refresh ttl %s exceeds max ttl %s", s.RefreshTTL, s.MaxTTL)
	}

	now := time.Now().UTC()

	aud := req.GetConfig().GetAudience()
	if len(aud) == 0 {
		aud = s.DefaultAud
	}
	if len(aud) == 0 {
		aud = []string{"multikube-proxy"}
	}

	accessClaims := jwt.MapClaims{
		"iss":          s.Issuer,
		"sub":          req.GetConfig().GetSubject(),
		"aud":          aud,
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
		"exp":          now.Add(accessTTL).Unix(),
		"jti":          uuid.NewString(),
		tokenTypeClaim: accessTokenType,
	}

	refreshClaims := jwt.MapClaims{
		"iss":          s.Issuer,
		"sub":          req.GetConfig().GetSubject(),
		"aud":          aud,
		"iat":          now.Unix(),
		"nbf":          now.Unix(),
		"exp":          now.Add(s.RefreshTTL).Unix(),
		"jti":          uuid.NewString(),
		tokenTypeClaim: refreshTokenType,
	}

	if req.GetConfig().GetUsername() != "" {
		accessClaims["preferred_username"] = req.GetConfig().GetUsername()
	}
	if len(req.GetConfig().GetGroups()) > 0 {
		accessClaims["groups"] = req.GetConfig().GetGroups()
	}
	if len(req.GetConfig().GetServiceAccounts()) > 0 {
		accessClaims["service_accounts"] = req.GetConfig().GetServiceAccounts()
	}
	if len(req.GetConfig().GetScopes()) > 0 {
		accessClaims["scope"] = req.GetConfig().GetScopes()
	}
	if len(req.GetConfig().GetClusters()) > 0 {
		accessClaims["clusters"] = req.GetConfig().GetClusters()
	}

	for k, v := range req.GetConfig().GetExtraClaims() {
		switch k {
		case "iss", "sub", "aud", "iat", "nbf", "exp", "jti", tokenTypeClaim, "groups", "service_accounts":
			return "", "", fmt.Errorf("extra claim %q is reserved", k)
		default:
			accessClaims[k] = v
		}
	}

	accessToken, err := s.generateSignedToken(accessClaims)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateSignedToken(refreshClaims)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *TokenManager) generateSignedToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.SigningKeyID
	token.Header["typ"] = "JWT"
	signed, err := token.SignedString(s.Key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (s *TokenManager) VerifyAccessToken(ctx context.Context, accessToken string) (string, error) {
	return s.verifyToken(ctx, accessToken, accessTokenType)
}

func (s *TokenManager) VerifyRefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return s.verifyToken(ctx, refreshToken, refreshTokenType)
}

func (s *TokenManager) verifyToken(ctx context.Context, rawToken string, expectedType string) (string, error) {
	_ = ctx

	token, err := jwt.ParseWithClaims(
		rawToken,
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
		return "", fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, err := claimsFromToken(token)
	if err != nil {
		return "", err
	}
	if err := s.validateClaims(claims, token, expectedType); err != nil {
		return "", err
	}

	sub, _ := claims["sub"].(string)
	return sub, nil
}

func claimsFromToken(token *jwt.Token) (jwt.MapClaims, error) {
	var claims jwt.MapClaims
	switch c := token.Claims.(type) {
	case jwt.MapClaims:
		claims = c
	case *jwt.MapClaims:
		claims = *c
	default:
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (s *TokenManager) validateClaims(claims jwt.MapClaims, token *jwt.Token, expectedType string) error {
	if typ, ok := token.Header["typ"].(string); ok && typ != "JWT" {
		return errors.New("invalid token type")
	}
	if !claims.VerifyIssuer(s.Issuer, true) {
		return errors.New("invalid token issuer")
	}
	if sub, ok := claims["sub"].(string); !ok || sub == "" {
		return errors.New("invalid token subject")
	}
	if typ, ok := claims[tokenTypeClaim].(string); !ok || typ != expectedType {
		return fmt.Errorf("invalid token purpose")
	}

	allowedAud := s.AllowedAud
	if len(allowedAud) == 0 {
		allowedAud = s.DefaultAud
	}
	if len(allowedAud) == 0 {
		return errors.New("allowed audiences are not configured")
	}

	var audValid bool
	for _, aud := range allowedAud {
		if claims.VerifyAudience(aud, true) {
			audValid = true
			break
		}
	}
	if !audValid {
		return errors.New("invalid token audience")
	}

	return nil
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
		AccessTTL:    10 * time.Minute,
		RefreshTTL:   24 * time.Hour,
		MaxTTL:       time.Hour * 72,
		DefaultAud:   []string{"multikube"},
		SigningKeyID: "key-2026-04",
	}

	for _, opt := range opts {
		opt(t)
	}

	return t, nil
}
