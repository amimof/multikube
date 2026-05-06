package main

import (
	"context"
	"testing"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
	tokenclientv1 "github.com/amimof/multikube/pkg/client/token/v1"
	"github.com/spf13/cobra"
)

type stubTokenClient struct {
	issued *tokenv1.Token
	resp   *tokenv1.IssueResponse
}

func (s *stubTokenClient) IssueToken(_ context.Context, tok *tokenv1.Token) (*tokenv1.IssueResponse, error) {
	s.issued = tok
	if s.resp != nil {
		return s.resp, nil
	}
	return &tokenv1.IssueResponse{AccessToken: "access-token"}, nil
}

var _ tokenclientv1.ClientV1 = (*stubTokenClient)(nil)

func TestRunCreateTokenCmdOmitsTTLWhenFlagNotProvided(t *testing.T) {
	stub := &stubTokenClient{}
	cleanup := setCreateTokenClientFactory(stub)
	defer cleanup()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.Flags().Uint64("ttl", 0, "")

	if err := runCreateTokenCmd(cmd, &cfg, "alice", "", "", nil, nil, nil, nil, 0, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.issued == nil || stub.issued.GetConfig() == nil {
		t.Fatal("expected token request to be issued")
	}
	if stub.issued.Config.Ttl != nil {
		t.Fatalf("ttl = %v, want nil", stub.issued.Config.Ttl)
	}
}

func TestRunCreateTokenCmdSetsZeroTTLWhenFlagProvided(t *testing.T) {
	stub := &stubTokenClient{}
	cleanup := setCreateTokenClientFactory(stub)
	defer cleanup()

	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.Flags().Uint64("ttl", 0, "")
	if err := cmd.Flags().Set("ttl", "0"); err != nil {
		t.Fatalf("set ttl flag: %v", err)
	}

	if err := runCreateTokenCmd(cmd, &cfg, "alice", "", "", nil, nil, nil, nil, 0, []string{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.issued == nil || stub.issued.GetConfig() == nil || stub.issued.Config.Ttl == nil {
		t.Fatal("expected ttl to be sent when flag is provided")
	}
	if got := stub.issued.GetConfig().GetTtl(); got != 0 {
		t.Fatalf("ttl = %d, want 0", got)
	}
}

func setCreateTokenClientFactory(tokenClient tokenclientv1.ClientV1) func() {
	original := createTokenClientFactory
	createTokenClientFactory = func() tokenclientv1.ClientV1 {
		return tokenClient
	}
	return func() {
		createTokenClientFactory = original
	}
}
