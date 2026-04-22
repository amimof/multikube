package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	auditv1 "github.com/amimof/multikube/api/audit/v1"
	"github.com/amimof/multikube/pkg/logger"
)

type Store interface {
	Subscriber
	Start(ctx context.Context) error
	Close() error
	All() []*auditv1.AuditEntry
	Range(from, to uint64) []*auditv1.AuditEntry
	From(seq uint64) []*auditv1.AuditEntry
}

type subscriber struct {
	id uint64
	ch chan *auditv1.AuditEntry
}

type FileStore struct {
	path         string
	mu           sync.RWMutex
	entries      []*auditv1.AuditEntry
	nextSeq      uint64
	pollInterval time.Duration
	fileMu       sync.Mutex
	file         *os.File
	offset       int64
	partial      []byte
	subMu        sync.Mutex
	subscribers  map[uint64]*subscriber
	nextSubID    uint64
	startOnce    sync.Once
	closed       atomic.Bool
	logger       logger.Logger
}

type FileStoreOption func(*FileStore)

func trimLine(b []byte) []byte {
	for len(b) > 0 {
		last := b[len(b)-1]
		if last == '\n' || last == '\r' || last == ' ' || last == '\t' {
			b = b[:len(b)-1]
			continue
		}
		break
	}
	return b
}

func WithFileStoreLogger(l logger.Logger) FileStoreOption {
	return func(fs *FileStore) {
		fs.logger = l
	}
}

func NewFileStore(path string, opts ...FileStoreOption) *FileStore {
	fs := &FileStore{
		path:         path,
		subscribers:  make(map[uint64]*subscriber),
		pollInterval: time.Second,
		logger:       logger.ConsoleLogger{},
	}

	for _, opt := range opts {
		opt(fs)
	}

	return fs
}

func (s *FileStore) Start(ctx context.Context) error {
	var startErr error

	s.startOnce.Do(func() {
		f, err := os.Open(s.path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				f, err = os.Create(s.path)
				if err != nil {
					startErr = fmt.Errorf("create audit file: %w", err)
					return
				}
				s.fileMu.Lock()
				s.file = f
				s.fileMu.Unlock()
			} else {
				startErr = fmt.Errorf("open audit file: %w", err)
				return
			}
		}

		s.fileMu.Lock()
		s.file = f
		s.fileMu.Unlock()

		// Initial catch-up.
		if err := s.readAvailable(); err != nil {
			startErr = err
			_ = f.Close()
			return
		}

		go s.watchLoop(ctx)
	})

	return startErr
}

func (s *FileStore) AppendLine(line []byte) error {
	if s.closed.Load() {
		return errors.New("service closed")
	}

	if !json.Valid(line) {
		return fmt.Errorf("invalid json line: %s", string(line))
	}

	var entry auditv1.AuditEntry
	err := json.Unmarshal(line, &entry)
	if err != nil {
		return err
	}

	seq := atomic.AddUint64(&s.nextSeq, 1) - 1
	entry.Seq = seq

	s.mu.Lock()
	s.entries = append(s.entries, &entry)
	s.mu.Unlock()

	s.broadcast(&entry)
	return nil
}

func (s *FileStore) All() []*auditv1.AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*auditv1.AuditEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

func (s *FileStore) From(seq uint64) []*auditv1.AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if seq >= uint64(len(s.entries)) {
		return nil
	}

	out := make([]*auditv1.AuditEntry, len(s.entries[seq:]))
	copy(out, s.entries[seq:])
	return out
}

func (s *FileStore) Range(from, to uint64) []*auditv1.AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n := uint64(len(s.entries))
	if from > n {
		from = n
	}
	if to > n {
		to = n
	}
	if from >= to {
		return nil
	}

	out := make([]*auditv1.AuditEntry, to-from)
	copy(out, s.entries[from:to])
	return out
}

func (s *FileStore) Subscribe(ctx context.Context, fromSeq uint64, live bool) (<-chan *auditv1.AuditEntry, error) {
	out := make(chan *auditv1.AuditEntry, 256)

	if !live {
		snapshot := s.From(fromSeq)
		go func() {
			defer close(out)
			for _, e := range snapshot {
				select {
				case <-ctx.Done():
					return
				case out <- e:
				}
			}
		}()
		return out, nil
	}
	sub := s.addSubscriber()
	snapshot := s.From(fromSeq)

	go func() {
		defer s.removeSubscriber(sub.id)

		var lastSeq uint64
		for _, e := range snapshot {
			select {
			case <-ctx.Done():
				return
			case out <- e:
				lastSeq = e.Seq
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-sub.ch:
				if !ok {
					close(out)
					return
				}
				if e.Seq <= lastSeq {
					continue
				}
				lastSeq = e.Seq
				select {
				case <-ctx.Done():
					return
				case out <- e:
				}
			}
		}
	}()
	return out, nil
}

func (s *FileStore) Publish(entry *auditv1.AuditEntry) {
}

func (s *FileStore) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}

	_ = s.file.Close()

	s.subMu.Lock()
	defer s.subMu.Unlock()

	for id, sub := range s.subscribers {
		close(sub.ch)
		delete(s.subscribers, id)
	}

	return nil
}

func (s *FileStore) readAvailable() error {
	s.fileMu.Lock()
	defer s.fileMu.Unlock()

	if s.file == nil {
		return errors.New("file not open")
	}

	if _, err := s.file.Seek(s.offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	buf := make([]byte, 64*1024)

	for {
		n, err := s.file.Read(buf)
		if n > 0 {
			chunk := append([]byte(nil), buf[:n]...)
			s.partial = append(s.partial, chunk...)

			for {
				idx := bytes.IndexByte(s.partial, '\n')
				if idx < 0 {
					break
				}

				line := trimLine(s.partial[:idx])
				s.partial = s.partial[idx+1:]

				if len(line) == 0 {
					continue
				}
				if err := s.AppendLine(line); err != nil {
					return err
				}
			}

			s.offset += int64(n)
		}

		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}
	}
}

func (s *FileStore) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.readAvailable(); err != nil {
				// Handle truncation / rotation before giving up.
				if recoverErr := s.handleFileChange(); recoverErr != nil {
					s.logger.Error("watch loop failed", "error", err)
					return
				}
			}
		}
	}
}

func (s *FileStore) handleFileChange() error {
	s.fileMu.Lock()
	defer s.fileMu.Unlock()

	if s.file != nil {
		info, err := s.file.Stat()
		if err == nil && s.offset <= info.Size() {
			return nil
		}
		_ = s.file.Close()
		s.file = nil
	}

	f, err := os.Open(s.path)
	if err != nil {
		return fmt.Errorf("reopen file: %w", err)
	}

	s.file = f
	s.offset = 0
	s.partial = nil

	return nil
}

func (s *FileStore) broadcast(entry *auditv1.AuditEntry) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for id, sub := range s.subscribers {
		select {
		case sub.ch <- entry:
		default:
			// Slow subscriber: disconnect it so one client cannot block everyone.
			close(sub.ch)
			delete(s.subscribers, id)
		}
	}
}

func (s *FileStore) addSubscriber() *subscriber {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	id := s.nextSubID
	s.nextSubID++

	sub := &subscriber{
		id: id,
		ch: make(chan *auditv1.AuditEntry, 256),
	}
	s.subscribers[id] = sub
	return sub
}

func (s *FileStore) removeSubscriber(id uint64) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	sub, ok := s.subscribers[id]
	if !ok {
		return
	}
	delete(s.subscribers, id)
	close(sub.ch)
}
