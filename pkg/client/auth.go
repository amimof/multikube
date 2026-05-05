package client

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	authclientv1 "github.com/amimof/multikube/pkg/client/auth/v1"
)

const refreshMethod = "/auth.v1.AuthService/Refresh"

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type TokenSource interface {
	Token(ctx context.Context) (*Token, error)
}

func (t *Token) GetAccessToken(_ context.Context) (string, bool) {
	return t.AccessToken, true
}

type RefreshTokenSource struct {
	mu sync.Mutex

	AuthClient authclientv1.ClientV1

	AccessToken  string
	RefreshToken string
	expiresAt    time.Time

	refreshBefore time.Duration
}

func NewRefreshTokenSource(authClient authclientv1.ClientV1, accessToken string, refreshToken string, expiresAt time.Time) *RefreshTokenSource {
	return &RefreshTokenSource{
		AuthClient:    authClient,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		expiresAt:     expiresAt,
		refreshBefore: time.Minute,
	}
}

func (s *RefreshTokenSource) Token(ctx context.Context) (*Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.AccessToken != "" && time.Until(s.expiresAt) > s.refreshBefore {
		return &Token{
			AccessToken:  s.AccessToken,
			RefreshToken: s.RefreshToken,
			ExpiresAt:    s.expiresAt,
		}, nil
	}
	if s.AccessToken != "" && s.expiresAt.IsZero() {
		return &Token{
			AccessToken:  s.AccessToken,
			RefreshToken: s.RefreshToken,
			ExpiresAt:    s.expiresAt,
		}, nil
	}

	resp, err := s.AuthClient.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: s.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	s.AccessToken = resp.AccessToken
	s.RefreshToken = resp.RefreshToken
	if resp.ExpiresAt != nil {
		s.expiresAt = resp.ExpiresAt.AsTime()
	} else {
		s.expiresAt = time.Time{}
	}

	return &Token{
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		ExpiresAt:    s.expiresAt,
	}, nil
}

func (s *RefreshTokenSource) GetAccessToken(_ context.Context) (string, bool) {
	if s == nil || s.AccessToken == "" {
		return "", false
	}

	return s.AccessToken, true
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

// AccessTokenUnaryInterceptor is a unary interceptor that adds access token metadata to outgoing context.
// Does not handle token refresh. Use [RefreshUnaryInterceptor] instead
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

// AccessTokenStreamInterceptor is a stream interceptor that adds access token metadata to outgoing context
// Does not handle token refresh. Use [RefreshStreamInterceptor] instead
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

func RefreshUnaryInterceptor(source *RefreshTokenSource) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if method == refreshMethod {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		err := invoker(ctx, method, req, reply, cc, opts...)
		if status.Code(err) != codes.Unauthenticated {
			return err
		}

		tok, err := source.Token(ctx)
		if err != nil {
			return err
		}

		ctx = withAccessToken(ctx, tok)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func RefreshStreamInterceptor(source *RefreshTokenSource) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if method == refreshMethod {
			return streamer(ctx, desc, cc, method, opts...)
		}

		cs, err := streamer(ctx, desc, cc, method, opts...)
		if status.Code(err) != codes.Unauthenticated {
			return cs, err
		}

		tok, err := source.Token(ctx)
		if err != nil {
			return nil, err
		}

		ctx = withAccessToken(ctx, tok)
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
