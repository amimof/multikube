package infra

import (
	"context"
	"errors"
	"io"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"

	eventv1 "github.com/amimof/multikube/api/event/v1"
)

var (
	ErrNodeNotConnected = errors.New("node not connected")
	ErrNodeQueueFull    = errors.New("node outbound queue full")
)

type SessionManager interface {
	Connect(ctx context.Context, in NodeConnectInput) (Session, error)
}

type Session interface {
	Handle(ctx context.Context, ev *eventv1.Envelope) error
	Next(ctx context.Context) (*eventv1.Envelope, error)
	Close() error
}

type NodeSessionManagerOption func(*NodeSessionManager)

// NodeConnectInput represents the identity established for a node stream.
// NodeUID and NodeName are expected to come from transport metadata.
type NodeConnectInput struct {
	NodeUID  string
	NodeName string
}

func WithOutboundBuffer(size int) NodeSessionManagerOption {
	return func(m *NodeSessionManager) {
		if size > 0 {
			m.outBuf = size
		}
	}
}

// NodeSessionManager owns node stream sessions and enables targeted delivery.
//
// Intended usage:
// - transport calls Connect() to open a session for a stream
// - transport forwards inbound stream messages to Session.Handle()
// - transport writes outbound messages from Session.Next() to the stream
// - business logic sends targeted events using SendToNode()
type NodeSessionManager struct {
	exchange *events.Exchange
	logger   logger.Logger

	mu       sync.Mutex
	sessions map[string]*nodeSession
	outBuf   int
}

type nodeSession struct {
	manager *NodeSessionManager

	nodeUID  string
	nodeName string
	// node     *nodesv1.Node

	out chan *eventv1.Envelope

	closeOnce sync.Once
	closed    chan struct{}
}

// func (m *NodeSessionManager) Subscribe(ctx context.Context, in NodeConnectInput, evType ...eventsv1.EventType) (Session, error) {
// 	eventChan := m.exchange.Subscribe(ctx, evType...)
// 	sess := &nodeSession{
// 		manager:  m,
// 		nodeUID:  in.NodeUID,
// 		nodeName: in.NodeName,
// 		out:      make(chan *eventsv1.Event, m.outBuf),
// 		closed:   make(chan struct{}),
// 	}
// }

func NewNodeSessionManager(exchange *events.Exchange, l logger.Logger, opts ...NodeSessionManagerOption) *NodeSessionManager {
	m := &NodeSessionManager{
		exchange: exchange,
		logger:   l,
		sessions: make(map[string]*nodeSession),
		outBuf:   64,
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.logger == nil {
		m.logger = logger.ConsoleLogger{}
	}
	return m
}

func (m *NodeSessionManager) Connect(ctx context.Context, in NodeConnectInput) (Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if in.NodeUID == "" {
		return nil, status.Error(codes.FailedPrecondition, "missing node uid")
	}
	if in.NodeName == "" {
		return nil, status.Error(codes.FailedPrecondition, "missing node name")
	}

	sess := &nodeSession{
		manager:  m,
		nodeUID:  in.NodeUID,
		nodeName: in.NodeName,
		// node:     node,
		out:    make(chan *eventv1.Envelope, m.outBuf),
		closed: make(chan struct{}),
	}

	var old *nodeSession
	m.mu.Lock()
	old = m.sessions[in.NodeUID]
	m.sessions[in.NodeUID] = sess
	m.mu.Unlock()

	// Close old session (reconnect) outside the lock.
	if old != nil {
		_ = old.Close()
	}

	// Publish NodeConnect.
	if m.exchange != nil {
		if err := m.exchange.Publish(ctx, events.NewEvent(events.CientConnect, nil)); err != nil {
			m.logger.Error("failed to publish node connect", "error", err, "nodeUID", in.NodeUID, "nodeName", in.NodeName)
		}
	}

	m.logger.Info("node connected", "nodeUID", in.NodeUID, "nodeName", in.NodeName)
	return sess, nil
}

func (m *NodeSessionManager) IsNodeConnected(nodeUID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.sessions[nodeUID]
	return ok
}

// Disconnect closes and removes the session for nodeUID, if present.
func (m *NodeSessionManager) Disconnect(nodeUID string) {
	m.mu.Lock()
	sess := m.sessions[nodeUID]
	m.mu.Unlock()
	if sess != nil {
		_ = sess.Close()
	}
}

func (s *nodeSession) Handle(ctx context.Context, ev *eventv1.Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if ev == nil {
		return status.Error(codes.InvalidArgument, "event is required")
	}
	if s.manager.exchange == nil {
		return status.Error(codes.FailedPrecondition, "event exchange is not configured")
	}
	return s.manager.exchange.Publish(ctx, ev)
}

func (s *nodeSession) Next(ctx context.Context) (*eventv1.Envelope, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.closed:
		return nil, io.EOF
	case ev, ok := <-s.out:
		if !ok {
			return nil, io.EOF
		}
		return ev, nil
	}
}

func (s *nodeSession) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)

		// Remove from manager map if we're still the active session.
		m := s.manager
		m.mu.Lock()
		cur := m.sessions[s.nodeUID]
		if cur == s {
			delete(m.sessions, s.nodeUID)
		}
		m.mu.Unlock()

		// Stop writers.
		close(s.out)

		// // Publish NodeForget.
		// if m.exchange != nil && s.node != nil {
		// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// 	defer cancel()
		// 	if err := m.exchange.Publish(ctx, events.NewEvent(events.NodeForget, s.node)); err != nil {
		// 		m.logger.Error("failed to publish node forget", "error", err, "nodeUID", s.nodeUID, "nodeName", s.nodeName)
		// 	}
		// }

		m.logger.Info("node disconnected", "nodeUID", s.nodeUID, "nodeName", s.nodeName)
	})
	return nil
}
