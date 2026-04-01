package app

import (
	"context"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
)

type TokenManager interface {
	Issue(ctx context.Context, token *tokenv1.Token) (string, error)
	Revoke(ctx context.Context, token *tokenv1.Token) error
}

type TokenValidator interface {
	Validate(ctx context.Context, rawToken string) (proxy.Principal, error)
}
