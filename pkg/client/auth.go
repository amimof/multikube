package client

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	authclientv1 "github.com/amimof/multikube/pkg/client/auth/v1"
)

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenSource interface {
	Token(ctx context.Context) (*Token, error)
}

type RefreshTokenSource struct {
	mu sync.Mutex

	authClient authclientv1.ClientV1

	accessToken  string
	refreshToken string
	expiresAt    time.Time

	refreshBefore time.Duration
}

func NewRefreshTokenSource(authClient authclientv1.ClientV1, accessToken string, refreshToken string, expiresAt time.Time) *RefreshTokenSource {
	return &RefreshTokenSource{
		authClient:    authClient,
		accessToken:   accessToken,
		refreshToken:  refreshToken,
		expiresAt:     expiresAt,
		refreshBefore: time.Minute,
	}
}

func (s *RefreshTokenSource) Token(ctx context.Context) (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.accessToken != "" && time.Until(s.expiresAt) > s.refreshBefore {
		return &Token{
			AccessToken:  s.accessToken,
			RefreshToken: s.refreshToken,
			ExpiresAt:    s.expiresAt,
		}, nil
	}
	if s.accessToken != "" && s.expiresAt.IsZero() {
		return &Token{
			AccessToken:  s.accessToken,
			RefreshToken: s.refreshToken,
			ExpiresAt:    s.expiresAt,
		}, nil
	}

	resp, err := s.authClient.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: s.refreshToken,
	})
	if err != nil {
		return nil, err
	}

	s.accessToken = resp.AccessToken
	s.refreshToken = resp.RefreshToken
	if resp.ExpiresAt != nil {
		s.expiresAt = resp.ExpiresAt.AsTime()
	} else {
		s.expiresAt = time.Time{}
	}

	return &Token{
		AccessToken:  s.accessToken,
		RefreshToken: s.refreshToken,
		ExpiresAt:    s.expiresAt,
	}, nil
}

type AccessTokenProvider interface {
	GetAccessToken(ctx context.Context) (string, bool)
}

type ConfigAccessTokenProvider struct {
	Config *Config
}

func (p ConfigAccessTokenProvider) GetAccessToken(context.Context) (string, bool) {
	if p.Config == nil {
		return "", false
	}

	server, err := p.Config.CurrentServer()
	if err != nil || server == nil || server.Session == nil || server.Session.AccessToken == "" {
		return "", false
	}

	return server.Session.AccessToken, true
}

func AccessTokenUnaryInterceptor(p AccessTokenProvider) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = withAccessToken(ctx, p)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AccessTokenStreamInterceptor(p AccessTokenProvider) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = withAccessToken(ctx, p)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func withAccessToken(ctx context.Context, p AccessTokenProvider) context.Context {
	if p == nil {
		return ctx
	}

	token, ok := p.GetAccessToken(ctx)
	if !ok || token == "" {
		return ctx
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)
}

type PerRPCTokenCredentials struct {
	source TokenSource
}

func (c *PerRPCTokenCredentials) GetRequestMetadata(
	ctx context.Context,
	uri ...string,
) (map[string]string, error) {
	token, err := c.source.Token(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"authorization": "Bearer " + token.AccessToken,
	}, nil
}

func (c *PerRPCTokenCredentials) RequireTransportSecurity() bool {
	return true
}
