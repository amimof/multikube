package app

import (
	"context"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

type TokenService struct {
	// Repo     *repository.Repo[*tokenv1.Token]
	Exchange *events.Exchange
	Logger   logger.Logger
	// Key          *ecdsa.PrivateKey
	// DefaultTTL   time.Duration
	// MaxTTL       time.Duration
	// DefaultAud   []string
	// Issuer       string
	// SigningKeyID string
	Issuer TokenManager
}

// type IssueResponse struct {
// 	AccessToken string
// 	ExpiresAt   time.Time
// 	KeyID       string
// 	TokenType   string
// }

func (s *TokenService) IssueToken(ctx context.Context, req *tokenv1.Token) (*tokenv1.IssueResponse, error) {
	_, span := tracer.Start(ctx, "token.Issue")
	defer span.End()

	accessToken, _, err := s.Issuer.Issue(ctx, req)
	if err != nil {
		return nil, err
	}

	// Publish event that token is issued
	err = s.Exchange.Forward(ctx, events.NewEvent(events.TokenIssue, nil))
	if err != nil {
		s.Logger.Error("error publishing token issuance event", "error", err, "subject", req.GetConfig().GetSubject())
		return nil, err
	}

	return &tokenv1.IssueResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *TokenService) Verify(ctx context.Context, accessToken string) error {
	_, span := tracer.Start(ctx, "token.Verify")
	defer span.End()

	_, err := s.Issuer.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		return err
	}
	return nil
}

// Revoke publishes a delete request and the subscribers are responsible for deleting resources.
// Once they do, they will update there resource with the status Deleted
func (s *TokenService) Revoke(ctx context.Context, id keys.ID) error {
	_, span := tracer.Start(ctx, "token.Revoke")
	defer span.End()

	// Publish event that token is revoked
	err := s.Exchange.Forward(ctx, events.NewEvent(events.TokenRevoke, nil))
	if err != nil {
		s.Logger.Error("error publishing token issuance event", "error", err, "keyID", id.String())
		return err
	}

	return nil
}
