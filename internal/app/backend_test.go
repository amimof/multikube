package app

import (
	"testing"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestApplyMaskedUpdateBackend_TargetStatusesPreserveSiblingBranch(t *testing.T) {
	dst := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			"http://example.com": {
				Readiness: &backendv1.TargetReadyStatus{
					IsReady:            boolPtr(true),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}
	src := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			"http://example.com": {
				Healthiness: &backendv1.TargetHealthStatus{
					IsHealthy:          boolPtr(false),
					Reason:             wrapperspb.String("boom"),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}

	if err := applyMaskedUpdateBackend(dst, src, &fieldmaskpb.FieldMask{Paths: []string{"target_statuses"}}); err != nil {
		t.Fatalf("apply masked update: %v", err)
	}

	got := dst.TargetStatuses["http://example.com"]
	if got.GetReadiness() == nil || !got.GetReadiness().GetIsReady() {
		t.Fatalf("expected readiness to be preserved, got %+v", got.GetReadiness())
	}
	if got.GetHealthiness() == nil || got.GetHealthiness().GetIsHealthy() {
		t.Fatalf("expected healthiness update to be applied, got %+v", got.GetHealthiness())
	}
}

func TestApplyMaskedUpdateBackend_TargetStatusesReplaceUpdatedSubtree(t *testing.T) {
	dst := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			"http://example.com": {
				Healthiness: &backendv1.TargetHealthStatus{
					IsHealthy:          boolPtr(false),
					Reason:             wrapperspb.String("previous failure"),
					LastTransitionTime: timestamppb.Now(),
				},
				Readiness: &backendv1.TargetReadyStatus{
					IsReady:            boolPtr(false),
					Reason:             wrapperspb.String("still warming up"),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}
	src := &backendv1.BackendStatus{
		TargetStatuses: map[string]*backendv1.TargetStatus{
			"http://example.com": {
				Healthiness: &backendv1.TargetHealthStatus{
					IsHealthy:          boolPtr(true),
					LastTransitionTime: timestamppb.Now(),
				},
			},
		},
	}

	if err := applyMaskedUpdateBackend(dst, src, &fieldmaskpb.FieldMask{Paths: []string{"target_statuses"}}); err != nil {
		t.Fatalf("apply masked update: %v", err)
	}

	got := dst.TargetStatuses["http://example.com"]
	if got.GetHealthiness() == nil || !got.GetHealthiness().GetIsHealthy() {
		t.Fatalf("expected healthy target status, got %+v", got.GetHealthiness())
	}
	if got.GetHealthiness().GetReason() != nil {
		t.Fatalf("expected health reason to be cleared, got %+v", got.GetHealthiness().GetReason())
	}
	if got.GetReadiness() == nil || got.GetReadiness().GetReason() == nil || got.GetReadiness().GetReason().GetValue() != "still warming up" {
		t.Fatalf("expected readiness subtree to remain untouched, got %+v", got.GetReadiness())
	}
}

func boolPtr(v bool) *bool { return &v }
