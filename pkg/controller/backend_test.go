package controller

import (
	"context"
	"sync"
	"testing"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	backendclientv1 "github.com/amimof/multikube/pkg/client/backend/v1"
	routeclientv1 "github.com/amimof/multikube/pkg/client/route/v1"
	"github.com/amimof/multikube/pkg/compile"
	"github.com/amimof/multikube/pkg/logger"
	proxyv2 "github.com/amimof/multikube/pkg/proxyv2"
	gomock "go.uber.org/mock/gomock"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mclient "github.com/amimof/multikube/pkg/client"
)

func boolPtr(v bool) *bool { return &v }

func TestControllerCompileRuntime_PublishesSnapshotAndStatuses(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// --- Route mock ---
	mockRouteService := routeclientv1.NewMockRouteServiceClient(mockCtrl)
	routeClient := routeclientv1.NewClientV1(routeclientv1.WithClient(mockRouteService))

	// --- Backend mock ---
	mockBackendService := backendclientv1.NewMockBackendServiceClient(mockCtrl)
	backendClient := backendclientv1.NewClientV1(backendclientv1.WithClient(mockBackendService))

	client, err := mclient.New("dummy",
		mclient.WithRouteClient(routeClient),
		mclient.WithBackendClient(backendClient),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Track route UpdateStatus calls.
	routeStatusCalls := map[string]*routev1.UpdateStatusRequest{}
	var statusCallsMu sync.Mutex
	recordRouteUpdate := func(_ context.Context, req *routev1.UpdateStatusRequest, _ ...any) (*emptypb.Empty, error) {
		statusCallsMu.Lock()
		defer statusCallsMu.Unlock()
		routeStatusCalls[req.GetName()] = req
		return &emptypb.Empty{}, nil
	}

	// Route UpdateStatus: 2 calls (for "invalid" and "valid").
	mockRouteService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(recordRouteUpdate).
		Times(2)

	// Route Get: 2 calls after each UpdateStatus to refresh the cache.
	// Return a route with the appropriate status based on the name.
	mockRouteService.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req *routev1.GetRequest, _ ...any) (*routev1.GetResponse, error) {
			name := req.GetName()
			phase := compile.PhaseReady
			reason := ""
			if name == "invalid" {
				phase = compile.PhaseInvalid
				reason = "route matcher is required"
			}
			return &routev1.GetResponse{
				Route: &routev1.Route{
					Meta: &metav1.Meta{Name: name},
					Config: &routev1.RouteConfig{
						BackendRef: "be",
						Enabled:    boolPtr(true),
					},
					Status: &routev1.RouteStatus{
						Phase:  wrapperspb.String(phase),
						Reason: wrapperspb.String(reason),
					},
				},
			}, nil
		}).
		Times(2)

	// Backend UpdateStatus: 1 call for "be" (PhaseReady).
	mockBackendService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req *backendv1.UpdateStatusRequest, _ ...any) (*emptypb.Empty, error) {
			return &emptypb.Empty{}, nil
		}).
		Times(1)

	// Backend Get: 1 call after UpdateStatus to refresh the cache.
	mockBackendService.EXPECT().
		Get(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, req *backendv1.GetRequest, _ ...any) (*backendv1.GetResponse, error) {
			return &backendv1.GetResponse{
				Backend: &backendv1.Backend{
					Meta: &metav1.Meta{Name: req.GetName()},
					Config: &backendv1.BackendConfig{
						Servers:               []string{"http://example.com"},
						InsecureSkipTlsVerify: true,
						Enabled:               boolPtr(true),
					},
					Status: &backendv1.BackendStatus{
						Phase: wrapperspb.String(compile.PhaseReady),
					},
				},
			}, nil
		}).
		Times(1)

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
						Enabled:               boolPtr(true),
					},
				},
			},
			Routes: map[string]*routev1.Route{
				"invalid": {
					Meta: &metav1.Meta{Name: "invalid"},
					Config: &routev1.RouteConfig{
						BackendRef: "be",
						Enabled:    boolPtr(true),
					},
				},
				"valid": {
					Meta: &metav1.Meta{Name: "valid"},
					Config: &routev1.RouteConfig{
						BackendRef: "be",
						Match:      &routev1.Match{Path: "/ok"},
						Enabled:    boolPtr(true),
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
	invalidReq := routeStatusCalls["invalid"]
	validReq := routeStatusCalls["valid"]
	statusCallsMu.Unlock()

	if invalidReq == nil || invalidReq.GetStatus().GetPhase().GetValue() != compile.PhaseInvalid {
		t.Fatalf("expected invalid route status update, got %+v", invalidReq)
	}
	if got := invalidReq.GetUpdateMask().GetPaths(); len(got) != 3 || got[0] != "phase" || got[1] != "reason" || got[2] != "last_transition_time" {
		t.Fatalf("expected status field mask, got %#v", got)
	}
	if ctrl.cache.Routes["invalid"].GetStatus().GetPhase().GetValue() != compile.PhaseInvalid {
		t.Fatalf("expected controller cache to refresh invalid route status, got %+v", ctrl.cache.Routes["invalid"].GetStatus())
	}
	if validReq == nil || validReq.GetStatus().GetPhase().GetValue() != compile.PhaseReady {
		t.Fatalf("expected ready route status update, got %+v", validReq)
	}
}
