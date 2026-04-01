package app

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

type TokenService struct {
	// Repo     *repository.Repo[*tokenv1.Token]
	Exchange     *events.Exchange
	Logger       logger.Logger
	Key          *ecdsa.PrivateKey
	DefaultTTL   time.Duration
	MaxTTL       time.Duration
	DefaultAud   []string
	Issuer       string
	SigningKeyID string
}

type IssueResponse struct {
	AccessToken string
	ExpiresAt   time.Time
	KeyID       string
	TokenType   string
}

func (s *TokenService) IssueToken(ctx context.Context, req *tokenv1.TokenConfig) (*tokenv1.IssueResponse, error) {
	_, span := tracer.Start(ctx, "token.Issue")
	defer span.End()

	if req.Subject == "" {
		return nil, errors.New("subject is required")
	}
	if req.Ttl == nil && s.DefaultTTL <= 0 {
		return nil, errors.New("default ttl must be configured")
	}
	if s.Key == nil {
		return nil, errors.New("signing key is not configured")
	}

	var ttl time.Duration
	if req.Ttl != nil {
		ttl = req.Ttl.AsDuration()
	}
	if ttl <= 0 {
		ttl = s.DefaultTTL
	}
	if ttl > s.MaxTTL {
		return nil, fmt.Errorf("requested ttl %s exceeds max ttl %s", ttl, s.MaxTTL)
	}

	now := time.Now().UTC()
	exp := now.Add(ttl)

	aud := req.Audience
	if len(aud) == 0 {
		aud = s.DefaultAud
	}
	if len(aud) == 0 {
		aud = []string{"multikube-proxy"}
	}

	claims := jwt.MapClaims{
		"iss": s.Issuer,
		"sub": req.Subject,
		"aud": aud,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"exp": exp.Unix(),
		"jti": uuid.NewString(),
	}

	if req.Username != "" {
		claims["preferred_username"] = req.Username
	}
	if len(req.Groups) > 0 {
		claims["groups"] = req.Groups
	}
	if len(req.ServiceAccounts) > 0 {
		claims["service_accounts"] = req.ServiceAccounts
	}
	if len(req.Scopes) > 0 {
		claims["scope"] = req.Scopes
	}
	if len(req.Clusters) > 0 {
		claims["clusters"] = req.Clusters
	}

	for k, v := range req.ExtraClaims {
		switch k {
		case "iss", "sub", "aud", "iat", "nbf", "exp", "jti", "groups", "service_accounts":
			return nil, fmt.Errorf("extra claim %q is reserved", k)
		default:
			claims[k] = v
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.SigningKeyID
	token.Header["typ"] = "JWT"

	signed, err := token.SignedString(s.Key)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &tokenv1.IssueResponse{
		AccessToken: signed,
		ExpiresAt:   timestamppb.New(exp),
		KeyId:       s.SigningKeyID,
		TokenType:   "Bearer",
	}, nil
}

// Revoke publishes a delete request and the subscribers are responsible for deleting resources.
// Once they do, they will update there resource with the status Deleted
func (s *TokenService) Revoke(ctx context.Context, id keys.ID) error {
	_, span := tracer.Start(ctx, "token.Revoke")
	defer span.End()

	// volume, err := l.Repo.Get(ctx, id)
	// if err != nil {
	// 	return err
	// }
	//
	// err = l.Repo.Delete(ctx, id)
	// if err != nil {
	// 	return err
	// }

	// err = l.Exchange.Forward(ctx, events.NewEvent(events.TokenDelete, volume))
	// if err != nil {
	// 	l.Logger.Error("error publishing volume delete event", "error", err, "name", volume.GetMeta().GetName())
	// 	return err
	// }

	return nil
}
