package controller

import (
	"context"
	"sync"
	"testing"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	routeclientv1 "github.com/amimof/multikube/pkg/client/route/v1"
	"github.com/amimof/multikube/pkg/compile"
	"github.com/amimof/multikube/pkg/logger"
	proxyv2 "github.com/amimof/multikube/pkg/proxyv2"
	gomock "go.uber.org/mock/gomock"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mclient "github.com/amimof/multikube/pkg/client"
)

func TestMergeRouteStatus_OnlyUpdatesTransitionOnPhaseChange(t *testing.T) {
	initial := time.Now().Add(-time.Hour)
	route := &routev1.Route{
		Status: &routev1.RouteStatus{
			Phase:              wrapperspb.String(compile.RoutePhaseInvalid),
			Reason:             wrapperspb.String("old"),
			LastTransitionTime: timestamppb.New(initial),
		},
	}

	updated, changed := mergeRouteStatus(route, compile.RouteCompileStatus{Phase: compile.RoutePhaseInvalid, Reason: "new"})
	if !changed {
		t.Fatal("expected status change when reason changes")
	}
	if got := updated.GetStatus().GetLastTransitionTime().AsTime(); !got.Equal(initial) {
		t.Fatalf("expected last transition time unchanged, got %v want %v", got, initial)
	}

	updated, changed = mergeRouteStatus(updated, compile.RouteCompileStatus{Phase: compile.RoutePhaseReady})
	if !changed {
		t.Fatal("expected status change when phase changes")
	}
	if !updated.GetStatus().GetLastTransitionTime().AsTime().After(initial) {
		t.Fatal("expected last transition time to advance on phase change")
	}
}

func TestControllerCompileRuntime_PublishesSnapshotAndStatuses(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRouteService := routeclientv1.NewMockRouteServiceClient(mockCtrl)
	routeClient := routeclientv1.NewClientV1(routeclientv1.WithClient(mockRouteService))

	client, err := mclient.New("dummy", mclient.WithRouteClient(routeClient))
	if err != nil {
		t.Fatal(err)
	}

	statusCalls := map[string]*routev1.UpdateStatusRequest{}
	var statusCallsMu sync.Mutex
	recordUpdate := func(_ context.Context, req *routev1.UpdateStatusRequest, _ ...any) (*emptypb.Empty, error) {
		statusCallsMu.Lock()
		defer statusCallsMu.Unlock()
		statusCalls[req.GetName()] = req
		return &emptypb.Empty{}, nil
	}

	mockRouteService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(recordUpdate).
		Times(2)

	runtimeStore := proxyv2.NewRuntimeStore()
	ctrl := &Controller{
		logger:    logger.NilLogger{},
		compiler:  compile.NewCompiler(),
		runtime:   runtimeStore,
		clientset: client,
		cache: &compile.State{
			Backends: map[string]*backendv1.Backend{
				"be": {
					Meta: &metav1.Meta{Name: "be"},
					Config: &backendv1.BackendConfig{
						Servers:               []string{"http://example.com"},
						InsecureSkipTlsVerify: true,
					},
				},
			},
			Routes: map[string]*routev1.Route{
				"invalid": {
					Meta: &metav1.Meta{Name: "invalid"},
					Config: &routev1.RouteConfig{
						BackendRef: "be",
					},
				},
				"valid": {
					Meta: &metav1.Meta{Name: "valid"},
					Config: &routev1.RouteConfig{
						BackendRef: "be",
						Match:      &routev1.Match{Path: "/ok"},
					},
				},
			},
			Certificates:           map[string]*certificatev1.Certificate{},
			CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
			Credentials:            map[string]*credentialv1.Credential{},
			Policies:               map[string]*policyv1.Policy{},
		},
	}

	if err := ctrl.compileRuntime(context.Background()); err != nil {
		t.Fatalf("compile runtime: %v", err)
	}

	rt := runtimeStore.Load()
	if len(rt.Routes.Paths) != 1 || rt.Routes.Paths[0].Name != "valid" {
		t.Fatalf("expected valid route to be published, got %+v", rt.Routes.Paths)
	}

	statusCallsMu.Lock()
	invalidReq := statusCalls["invalid"]
	validReq := statusCalls["valid"]
	statusCallsMu.Unlock()

	if invalidReq == nil || invalidReq.GetStatus().GetPhase().GetValue() != compile.RoutePhaseInvalid {
		t.Fatalf("expected invalid route status update, got %+v", invalidReq)
	}
	if got := invalidReq.GetUpdateMask().GetPaths(); len(got) != 3 || got[0] != "phase" || got[1] != "reason" || got[2] != "last_transition_time" {
		t.Fatalf("expected status field mask, got %#v", got)
	}
	if ctrl.cache.Routes["invalid"].GetStatus().GetPhase().GetValue() != compile.RoutePhaseInvalid {
		t.Fatalf("expected controller cache to refresh invalid route status, got %+v", ctrl.cache.Routes["invalid"].GetStatus())
	}
	if validReq == nil || validReq.GetStatus().GetPhase().GetValue() != compile.RoutePhaseReady {
		t.Fatalf("expected ready route status update, got %+v", validReq)
	}
}
