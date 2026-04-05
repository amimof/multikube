package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Sink interface {
	WriteBatch(ctx context.Context, events []*AuditEvent) error
	Close() error
}

type FileSink struct {
	mu sync.Mutex
	f  *os.File
	w  *bufio.Writer
}

func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}

	return &FileSink{
		f: f,
		w: bufio.NewWriterSize(f, 64*1024),
	}, nil
}

func (s *FileSink) WriteBatch(ctx context.Context, events []*AuditEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ev := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		b, err := json.Marshal(ev)
		if err != nil {
			return fmt.Errorf("marshal audit event: %w", err)
		}
		if _, err := s.w.Write(b); err != nil {
			return fmt.Errorf("write audit event: %w", err)
		}
		if err := s.w.WriteByte('\n'); err != nil {
			return fmt.Errorf("write newline: %w", err)
		}
	}

	if err := s.w.Flush(); err != nil {
		return fmt.Errorf("flush audit writer: %w", err)
	}

	return nil
}

func (s *FileSink) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.w != nil {
		if err := s.w.Flush(); err != nil {
			_ = s.f.Close()
			return err
		}
	}
	if s.f != nil {
		return s.f.Close()
	}
	return nil
}
