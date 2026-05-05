package infra

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
	"github.com/amimof/multikube/pkg/auth"
)

type NewTokenManagerOption func(*TokenManager)

type TokenManager struct {
	Key          *ecdsa.PrivateKey
	VerifyKey    *ecdsa.PublicKey
	AllowedAud   string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	MaxTTL       time.Duration
	DefaultAud   string
	Issuer       string
	SigningKeyID string
	required     map[string]auth.Permission
}

const (
	tokenTypeClaim   = "typ"
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

const nonExpiringTTL uint64 = 0

type ContextKey string

var (
	ContextKeyUsername ContextKey = "username"
	ContextKeyRoles    ContextKey = "roles"
)

type IssueResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	KeyID        string
	TokenType    string
}

type UserClaims struct {
	jwt.StandardClaims
	Username        string   `json:"preferred_username,omitempty"`
	Roles           []string `json:"roles,omitempty"`
	Groups          []string `json:"groups,omitempty"`
	Scopes          []string `json:"scopes,omitempty"`
	Clusters        []string `json:"clusters,omitempty"`
	ServiceAccounts []string `json:"service_accounts,omitempty"`
	Type            string   `json:"typ,omitempty"`
}

func (s *TokenManager) Issue(ctx context.Context, req *tokenv1.Token) (IssueResponse, error) {
	if req.GetConfig().GetSubject() == "" {
		return IssueResponse{}, errors.New("subject is required")
	}
	if s.AccessTTL <= 0 {
		return IssueResponse{}, errors.New("access ttl must be configured")
	}
	if s.RefreshTTL <= 0 {
		return IssueResponse{}, errors.New("refresh ttl must be configured")
	}
	if s.Key == nil {
		return IssueResponse{}, errors.New("signing key is not configured")
	}

	accessTTL := s.AccessTTL
	accessNeverExpires := false
	if req.GetConfig() != nil && req.GetConfig().Ttl != nil {
		if req.GetConfig().GetTtl() == nonExpiringTTL {
			accessNeverExpires = true
		} else {
			accessTTL = time.Duration(req.GetConfig().GetTtl()) * time.Second
		}
	}
	if !accessNeverExpires && accessTTL > s.MaxTTL {
		return IssueResponse{}, fmt.Errorf("requested ttl %s exceeds max ttl %s", accessTTL, s.MaxTTL)
	}
	if s.RefreshTTL > s.MaxTTL {
		return IssueResponse{}, fmt.Errorf("refresh ttl %s exceeds max ttl %s", s.RefreshTTL, s.MaxTTL)
	}

	now := time.Now().UTC()

	aud := req.GetConfig().GetAudience()
	if len(aud) == 0 {
		aud = s.DefaultAud
	}
	if len(aud) == 0 {
		aud = "multikube-proxy"
	}

	accessClaims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    s.Issuer,
			Subject:   req.GetConfig().GetSubject(),
			Audience:  aud,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			Id:        uuid.NewString(),
		},
		Type: accessTokenType,
	}
	if !accessNeverExpires {
		accessClaims.ExpiresAt = now.Add(accessTTL).Unix()
	}

	refreshClaims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    s.Issuer,
			Subject:   req.GetConfig().GetSubject(),
			Audience:  aud,
			IssuedAt:  now.Unix(),
			NotBefore: now.Unix(),
			ExpiresAt: now.Add(s.RefreshTTL).Unix(),
			Id:        uuid.NewString(),
		},
		Type: refreshTokenType,
	}

	if req.GetConfig().GetUsername() != "" {
		accessClaims.Username = req.GetConfig().GetUsername()
	}
	if len(req.GetConfig().GetGroups()) > 0 {
		accessClaims.Groups = req.GetConfig().GetGroups()
	}
	if len(req.GetConfig().GetServiceAccounts()) > 0 {
		accessClaims.ServiceAccounts = req.GetConfig().GetServiceAccounts()
	}
	if len(req.GetConfig().GetScopes()) > 0 {
		accessClaims.Scopes = req.GetConfig().GetScopes()
	}
	if len(req.GetConfig().GetClusters()) > 0 {
		accessClaims.Clusters = req.GetConfig().GetClusters()
	}
	if len(req.GetConfig().GetRoles()) > 0 {
		accessClaims.Roles = req.GetConfig().GetRoles()
	}

	extraClaims := make(map[string]string, len(req.GetConfig().GetExtraClaims()))
	for k, v := range req.GetConfig().GetExtraClaims() {
		switch k {
		case "iss", "sub", "aud", "iat", "nbf", "exp", "jti", tokenTypeClaim, "groups", "service_accounts", "roles":
			return IssueResponse{}, fmt.Errorf("extra claim %q is reserved", k)
		default:
			extraClaims[k] = v
		}
	}

	accessToken, err := s.generateSignedToken(accessClaims, extraClaims)
	if err != nil {
		return IssueResponse{}, err
	}

	refreshToken, err := s.generateSignedToken(refreshClaims, nil)
	if err != nil {
		return IssueResponse{}, err
	}

	response := IssueResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		KeyID:        accessClaims.Id,
		TokenType:    accessTokenType,
	}
	if !accessNeverExpires {
		response.ExpiresAt = time.Unix(accessClaims.ExpiresAt, 0)
	}

	return response, nil
}

func (s *TokenManager) generateSignedToken(claims UserClaims, extraClaims map[string]string) (string, error) {
	claimsMap := jwt.MapClaims{
		"iss": claims.Issuer,
		"sub": claims.Subject,
		"aud": claims.Audience,
		"iat": claims.IssuedAt,
		"nbf": claims.NotBefore,
		"jti": claims.Id,
		"typ": claims.Type,
	}
	if claims.ExpiresAt != 0 {
		claimsMap["exp"] = claims.ExpiresAt
	}
	if claims.Username != "" {
		claimsMap["preferred_username"] = claims.Username
	}
	if len(claims.Roles) > 0 {
		claimsMap["roles"] = claims.Roles
	}
	if len(claims.Groups) > 0 {
		claimsMap["groups"] = claims.Groups
	}
	if len(claims.Scopes) > 0 {
		claimsMap["scope"] = claims.Scopes
	}
	if len(claims.Clusters) > 0 {
		claimsMap["clusters"] = claims.Clusters
	}
	if len(claims.ServiceAccounts) > 0 {
		claimsMap["service_accounts"] = claims.ServiceAccounts
	}
	for k, v := range extraClaims {
		claimsMap[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claimsMap)
	token.Header["kid"] = s.SigningKeyID
	token.Header["typ"] = "JWT"
	signed, err := token.SignedString(s.Key)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (s *TokenManager) VerifyAccessToken(ctx context.Context, accessToken string) (jwt.MapClaims, error) {
	return s.verifyToken(ctx, accessToken, accessTokenType)
}

func (s *TokenManager) VerifyRefreshToken(ctx context.Context, refreshToken string) (jwt.MapClaims, error) {
	return s.verifyToken(ctx, refreshToken, refreshTokenType)
}

func (s *TokenManager) verifyToken(ctx context.Context, rawToken string, expectedType string) (jwt.MapClaims, error) {
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
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, err := claimsFromToken(token)
	if err != nil {
		return nil, err
	}
	if err := s.validateClaims(claims, token, expectedType); err != nil {
		return nil, err
	}

	// sub, _ := claims["sub"].(string)
	return claims, nil
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

	audValid := claims.VerifyAudience(allowedAud, true)

	if !audValid {
		return errors.New("invalid token audience")
	}

	return nil
}

func (s *TokenManager) Revoke(ctx context.Context, token *tokenv1.Token) error {
	return nil
}

func (s *TokenManager) AuthzUnaryInterceptor(required map[string]auth.Permission) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		permission, ok := required[info.FullMethod]
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "no permission rule for %s", info.FullMethod)
		}

		username, roles, err := s.UserFromContext(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing user")
		}
		if username == "" {
			return nil, status.Error(codes.Unauthenticated, "missing user")
		}

		if !auth.HasPermission(roles, permission) {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		return handler(ctx, req)
	}
}

func (s *TokenManager) UserFromContext(ctx context.Context) (string, []string, error) {
	accessToken, err := accessTokenFromContext(ctx)
	if err != nil {
		return "", nil, err
	}

	claims, err := s.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		return "", nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	username, _ := claims["sub"].(string)
	roles := stringSliceClaim(claims, "roles")
	return username, roles, nil
}

func (s *TokenManager) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if isReflectionMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		if auth.IsPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		claims, err := s.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		if claims != nil {
			username, _ := claims["sub"].(string)
			roles := stringSliceClaim(claims, "roles")

			ctx = context.WithValue(ctx, ContextKeyUsername, username)
			ctx = context.WithValue(ctx, ContextKeyRoles, roles)
		}

		return handler(ctx, req)
	}
}

func (s *TokenManager) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if isReflectionMethod(info.FullMethod) {
			return handler(srv, stream)
		}

		if auth.IsPublicMethod(info.FullMethod) {
			return handler(srv, stream)
		}

		ctx := stream.Context()
		claims, err := s.authorize(ctx, info.FullMethod)
		if err != nil {
			return err
		}
		if claims != nil {

			username, _ := claims["sub"].(string)
			roles := stringSliceClaim(claims, "roles")

			ctx = context.WithValue(ctx, ContextKeyUsername, username)
			ctx = context.WithValue(ctx, ContextKeyRoles, roles)
		}

		return handler(srv, &contextServerStream{ServerStream: stream, ctx: ctx})
	}
}

func (s *TokenManager) authorize(ctx context.Context, method string) (jwt.MapClaims, error) {
	permission, ok := s.required[method]
	if !ok {
		return nil, fmt.Errorf("method not allowed %s", method)
	}

	accessToken, err := accessTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	claims, err := s.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	roles := stringSliceClaim(claims, "roles")
	subject, _ := claims["sub"].(string)

	if !auth.HasPermission(roles, permission) {
		return nil, status.Errorf(codes.PermissionDenied, "user %s is not allowed to call %s", subject, method)
	}

	return claims, nil
}

type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextServerStream) Context() context.Context {
	return s.ctx
}

func stringSliceClaim(claims jwt.MapClaims, key string) []string {
	value, ok := claims[key]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []string:
		return typed
	case []interface{}:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			str, ok := item.(string)
			if !ok {
				return nil
			}
			values = append(values, str)
		}
		return values
	default:
		return nil
	}
}

func accessTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	token := strings.TrimSpace(values[0])
	if token == "" {
		return "", status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	parts := strings.Fields(token)
	if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
		return parts[1], nil
	}

	return token, nil
}

func isReflectionMethod(fullMethod string) bool {
	return fullMethod == "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo" ||
		fullMethod == "/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo"
}

func NewTokenManager(key *ecdsa.PrivateKey, opts ...NewTokenManagerOption) (*TokenManager, error) {
	t := &TokenManager{
		VerifyKey:    &key.PublicKey,
		AllowedAud:   "multikube",
		Issuer:       "https://auth.multikube.io",
		Key:          key,
		AccessTTL:    10 * time.Minute,
		RefreshTTL:   24 * time.Hour,
		MaxTTL:       time.Hour * 72,
		DefaultAud:   "multikube",
		SigningKeyID: "key-2026-04",
		required:     auth.PrivateMethods,
	}

	for _, opt := range opts {
		opt(t)
	}

	return t, nil
}
