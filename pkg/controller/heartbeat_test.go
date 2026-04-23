package controller

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	mclient "github.com/amimof/multikube/pkg/client"
	backendclientv1 "github.com/amimof/multikube/pkg/client/backend/v1"
	"github.com/amimof/multikube/pkg/logger"
	proxyv2 "github.com/amimof/multikube/pkg/proxyv2"
	gomock "go.uber.org/mock/gomock"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHeartbeatNext_ReturnsErrorOnNonOK(t *testing.T) {
	targetURL, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	hb := &Heartbeat{
		Runtime: &proxyv2.BackendRuntime{
			URL: targetURL,
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(http.NoBody),
				}, nil
			}),
		},
		Path:    "/healthz",
		Timeout: time.Second,
		Logger:  logger.NilLogger{},
	}

	err = hb.Next(context.Background())
	if err == nil {
		t.Fatal("expected non-200 heartbeat to return error")
	}
}

func TestHeartbeatStop_IsIdempotent(t *testing.T) {
	hb := &Heartbeat{stopCh: make(chan struct{})}

	done := make(chan struct{})
	go func() {
		hb.Stop()
		hb.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("stop should not block")
	}
}

func TestRemoveHealthProbe_RemovesRegisteredProbes(t *testing.T) {
	probeA := &Heartbeat{stopCh: make(chan struct{})}
	probeB := &Heartbeat{stopCh: make(chan struct{})}
	ctrl := &Controller{
		probes: map[string]map[string]*Heartbeat{
			"be": {
				"http://one": probeA,
				"http://two": probeB,
			},
		},
	}
	be := &backendv1.Backend{Meta: &metav1.Meta{Name: "be"}}

	ctrl.removeHealthProbe(context.Background(), be)

	if _, ok := ctrl.probes["be"]; ok {
		t.Fatal("expected backend probes to be removed")
	}

	select {
	case <-probeA.stopCh:
	default:
		t.Fatal("expected first probe to be stopped")
	}

	select {
	case <-probeB.stopCh:
	default:
		t.Fatal("expected second probe to be stopped")
	}
}

func TestRunHealthProbe_RegistersHealthAndReadyForSameTarget(t *testing.T) {
	backendName := "be"
	targetURL := mustParseURL(t, "http://example.com")
	ctrl := &Controller{
		logger:            logger.NilLogger{},
		heartBeatInterval: time.Second,
		heartBeatTimeout:  time.Second,
		probes:            map[string]map[string]*Heartbeat{},
		runtime:           proxyv2.NewRuntimeStore(),
	}
	ctrl.runtime.Store(&proxyv2.RuntimeConfig{
		Backends: map[string]*proxyv2.BackendPool{
			backendName: {
				Name: backendName,
				Targets: []*proxyv2.BackendRuntime{{
					Name: backendName,
					URL:  targetURL,
					Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
						return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(http.NoBody)}, nil
					}),
				}},
			},
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	health := &backendv1.Probe{Path: "/healthz", TimeoutSeconds: 1, PeriodSeconds: 3600, SuccessThreshold: 1, FailureThreshold: 1}
	ready := &backendv1.Probe{Path: "/readyz", TimeoutSeconds: 1, PeriodSeconds: 3600, SuccessThreshold: 1, FailureThreshold: 1}

	ctrl.runHealthProbe(ctx, backendName, health, &Callbacks{OnSuccess: func(*proxyv2.BackendRuntime) error { return nil }, OnFailure: func(*proxyv2.BackendRuntime, error) error { return nil }}, "health")
	ctrl.runHealthProbe(ctx, backendName, ready, &Callbacks{OnSuccess: func(*proxyv2.BackendRuntime) error { return nil }, OnFailure: func(*proxyv2.BackendRuntime, error) error { return nil }}, "ready")

	probes := ctrl.probes[backendName]
	if len(probes) != 2 {
		t.Fatalf("expected two probes for backend, got %d", len(probes))
	}

	var sawHealth bool
	var sawReady bool
	for key, hb := range probes {
		sawHealth = sawHealth || strings.HasPrefix(key, "health:") || hb.Kind == "health"
		sawReady = sawReady || strings.HasPrefix(key, "ready:") || hb.Kind == "ready"
	}
	if !sawHealth || !sawReady {
		t.Fatalf("expected both health and ready probes, got keys %#v", mapsKeys(probes))
	}

	ctrl.removeHealthProbe(ctx, &backendv1.Backend{Meta: &metav1.Meta{Name: backendName}})
}

func TestRunSingleHeartbeat_ResetsCountersOnFailureAndSuccess(t *testing.T) {
	ctrl := &Controller{logger: logger.NilLogger{}}
	target := &proxyv2.BackendRuntime{Name: "target"}

	failing := &Heartbeat{
		FailureThreshold: 100,
		SuccessThreshold: 100,
		Logger:           logger.NilLogger{},
		Runtime: &proxyv2.BackendRuntime{
			URL: mustParseURL(t, "http://example.com"),
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(http.NoBody),
				}, nil
			}),
		},
		Path:    "/healthz",
		Timeout: time.Second,
	}
	failing.SuccessCount.Store(3)

	ctrl.runSingleHeartbeat(context.Background(), "be", target, failing)
	if got := failing.SuccessCount.Load(); got != 0 {
		t.Fatalf("expected success counter reset on failure, got %d", got)
	}
	if got := failing.FailureCount.Load(); got != 1 {
		t.Fatalf("expected failure counter incremented on failure, got %d", got)
	}

	succeeding := &Heartbeat{
		FailureThreshold: 100,
		SuccessThreshold: 100,
		Logger:           logger.NilLogger{},
		Runtime: &proxyv2.BackendRuntime{
			URL: mustParseURL(t, "http://example.com"),
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(http.NoBody),
				}, nil
			}),
		},
		Path:    "/healthz",
		Timeout: time.Second,
	}
	succeeding.FailureCount.Store(4)

	ctrl.runSingleHeartbeat(context.Background(), "be", target, succeeding)
	if got := succeeding.FailureCount.Load(); got != 0 {
		t.Fatalf("expected failure counter reset on success, got %d", got)
	}
	if got := succeeding.SuccessCount.Load(); got != 1 {
		t.Fatalf("expected success counter incremented on success, got %d", got)
	}
}

func TestWaitForHeartbeat_RespectsDelay(t *testing.T) {
	start := time.Now()
	if !waitForHeartbeat(context.Background(), make(chan struct{}), 20*time.Millisecond) {
		t.Fatal("expected wait to complete")
	}
	if elapsed := time.Since(start); elapsed < 20*time.Millisecond {
		t.Fatalf("expected wait to honor delay, got %v", elapsed)
	}
}

func TestSetTargetUnhealthy_UpdatesLiveRuntimeTargetState(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBackendService := backendclientv1.NewMockBackendServiceClient(mockCtrl)
	mockBackendService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)
	client, err := mclient.New("dummy", mclient.WithBackendClient(backendclientv1.NewClientV1(backendclientv1.WithClient(mockBackendService))))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	target := &proxyv2.BackendRuntime{
		Name:           "be",
		URL:            mustParseURL(t, "http://example.com"),
		HasHealthProbe: true,
	}
	runtimeStore := proxyv2.NewRuntimeStore()
	runtimeStore.Store(&proxyv2.RuntimeConfig{Backends: map[string]*proxyv2.BackendPool{"be": {Name: "be", Targets: []*proxyv2.BackendRuntime{target}}}})
	ctrl := &Controller{logger: logger.NilLogger{}, clientset: client, runtime: runtimeStore, heartBeatTimeout: time.Second}

	if err := ctrl.setTargetUnhealthy(target, io.EOF); err != nil {
		t.Fatalf("set target unhealthy: %v", err)
	}
	if !target.HealthKnown.Load() || target.Healthy.Load() {
		t.Fatalf("expected runtime target to be marked unhealthy, got known=%v healthy=%v", target.HealthKnown.Load(), target.Healthy.Load())
	}
}

func TestSetTargetReady_UpdatesLatestRuntimeSnapshot(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockBackendService := backendclientv1.NewMockBackendServiceClient(mockCtrl)
	mockBackendService.EXPECT().
		UpdateStatus(gomock.Any(), gomock.Any()).
		Return(&emptypb.Empty{}, nil)
	client, err := mclient.New("dummy", mclient.WithBackendClient(backendclientv1.NewClientV1(backendclientv1.WithClient(mockBackendService))))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	stale := &proxyv2.BackendRuntime{Name: "be", URL: mustParseURL(t, "http://example.com"), HasReadinessProbe: true}
	latest := &proxyv2.BackendRuntime{Name: "be", URL: mustParseURL(t, "http://example.com"), HasReadinessProbe: true}
	runtimeStore := proxyv2.NewRuntimeStore()
	runtimeStore.Store(&proxyv2.RuntimeConfig{Version: 1, Backends: map[string]*proxyv2.BackendPool{"be": {Name: "be", Targets: []*proxyv2.BackendRuntime{stale}}}})
	runtimeStore.Store(&proxyv2.RuntimeConfig{Version: 2, Backends: map[string]*proxyv2.BackendPool{"be": {Name: "be", Targets: []*proxyv2.BackendRuntime{latest}}}})
	ctrl := &Controller{logger: logger.NilLogger{}, clientset: client, runtime: runtimeStore, heartBeatTimeout: time.Second}

	if err := ctrl.setTargetReady(stale); err != nil {
		t.Fatalf("set target ready: %v", err)
	}
	if !latest.ReadinessKnown.Load() || !latest.Ready.Load() {
		t.Fatalf("expected latest runtime target to be marked ready, got known=%v ready=%v", latest.ReadinessKnown.Load(), latest.Ready.Load())
	}
	if stale.ReadinessKnown.Load() || stale.Ready.Load() {
		t.Fatalf("expected stale heartbeat pointer to remain untouched, got known=%v ready=%v", stale.ReadinessKnown.Load(), stale.Ready.Load())
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u
}

func mapsKeys(m map[string]*Heartbeat) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
