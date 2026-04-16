package audit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	auditv1 "github.com/amimof/multikube/api/audit/v1"
	"github.com/amimof/multikube/pkg/logger"
)

type NewPublisherOpts func(*AsyncPublisher)

func WithLogger(logger logger.Logger) NewPublisherOpts {
	return func(ap *AsyncPublisher) {
		ap.logger = logger
	}
}

type Publisher interface {
	Publish(*auditv1.AuditEntry)
}

type AsyncPublisher struct {
	QueueSize      int
	BatchSize      int
	FlushInterval  time.Duration
	WriteTimeout   time.Duration
	OverflowPolicy OverflowPolicy

	sink      Sink
	logger    logger.Logger
	ch        chan *auditv1.AuditEntry
	stopCh    chan struct{}
	doneCh    chan struct{}
	closeOnce sync.Once
	stats     Stats
	running   atomic.Bool
}

func NewAsyncPublisher(sink Sink, opts ...NewPublisherOpts) *AsyncPublisher {
	return &AsyncPublisher{
		QueueSize:      4096,
		BatchSize:      256,
		FlushInterval:  250 * time.Millisecond,
		WriteTimeout:   5 * time.Second,
		OverflowPolicy: OverflowDrop,
		sink:           sink,
		logger:         logger.ConsoleLogger{},
		ch:             make(chan *auditv1.AuditEntry, 4096),
		stopCh:         make(chan struct{}),
		doneCh:         make(chan struct{}),
	}
}

func (p *AsyncPublisher) Publish(ev *auditv1.AuditEntry) {
	if ev == nil {
		return
	}

	p.stats.Published.Add(1)

	switch p.OverflowPolicy {
	case OverflowBlock:
		select {
		case p.ch <- ev:
		case <-p.stopCh:
			p.stats.Dropped.Add(1)
		}
	default:
		select {
		case p.ch <- ev:
		default:
			p.stats.Dropped.Add(1)
		}
	}
}

func (p *AsyncPublisher) Shutdown(ctx context.Context) error {
	var err error

	p.closeOnce.Do(func() {
		close(p.stopCh)

		select {
		case <-p.doneCh:
		case <-ctx.Done():
			err = ctx.Err()
		}

		close(p.ch)

		if sinkErr := p.sink.Close(); sinkErr != nil && err == nil {
			err = sinkErr
		}
	})

	return err
}

func (p *AsyncPublisher) run() {
	defer close(p.doneCh)

	ticker := time.NewTicker(p.FlushInterval)
	defer ticker.Stop()

	batch := make([]*auditv1.AuditEntry, 0, p.BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		p.flush(batch)
		batch = batch[:0]
	}

	for {
		select {
		case ev := <-p.ch:
			if ev != nil {
				batch = append(batch, ev)
				if len(batch) >= p.BatchSize {
					flush()
				}
			}

		case <-ticker.C:
			flush()

		case <-p.stopCh:
			// Drain remaining events quickly.
			for {
				select {
				case ev := <-p.ch:
					if ev != nil {
						batch = append(batch, ev)
						if len(batch) >= p.BatchSize {
							flush()
						}
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (p *AsyncPublisher) flush(events []*auditv1.AuditEntry) {
	p.stats.Flushes.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), p.WriteTimeout)
	defer cancel()

	if err := p.sink.WriteBatch(ctx, events); err != nil {
		p.stats.Failed.Add(uint64(len(events)))
		p.logger.Error("failed to write audit batch",
			"count", len(events),
			"error", err,
		)
		return
	}

	p.stats.Written.Add(uint64(len(events)))
}

func (p *AsyncPublisher) Start() {
	if p.running.CompareAndSwap(false, true) {
		go p.run()
	}
}
