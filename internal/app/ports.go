package app

import (
	"context"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
	userv1 "github.com/amimof/multikube/api/user/v1"
	"github.com/amimof/multikube/pkg/keys"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
	"github.com/golang-jwt/jwt"
)

type TokenManager interface {
	Issue(ctx context.Context, token *tokenv1.Token) (string, error)
	Revoke(ctx context.Context, token *tokenv1.Token) error
	Verify(ctx context.Context, accessToken string) (*jwt.Token, error)
}

type TokenValidator interface {
	Validate(ctx context.Context, rawToken string) (proxy.Principal, error)
}

type JWTManager interface {
	GenerateAccessToken(user *userv1.User) (string, error)
	GenerateRefreshToken(user *userv1.User) (string, error)
	Verify(accessToken, key string) (*UserClaims, error)
	VerifyAccessToken(token string) (*UserClaims, error)
	VerifyRefreshToken(token string) (*UserClaims, error)
}

type UserClaims struct {
	jwt.StandardClaims
	Username string   `json:"username"`
	ID       string   `json:"id"`
	Roles    []string `json:"roles"`
}

type UsersGetter interface {
	List(context.Context, int32) ([]*userv1.User, error)
	Get(context.Context, keys.ID) (*userv1.User, error)
}
