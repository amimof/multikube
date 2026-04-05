package proxy

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/amimof/multikube/pkg/audit"
)

type auditEventKey struct{}

func WithEvent(ctx context.Context, ev *audit.AuditEvent) context.Context {
	return context.WithValue(ctx, auditEventKey{}, ev)
}

func EventFromContext(ctx context.Context) (*audit.AuditEvent, bool) {
	ev, ok := ctx.Value(auditEventKey{}).(*audit.AuditEvent)
	return ev, ok
}

type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int64
}

func (w *auditResponseWriter) WriterHeader(code int) {
	w.statusCode = code
	w.WriteHeader(code)
}

func (w *auditResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	if err != nil {
		return 0, err
	}
	w.bytes += int64(n)
	return n, nil
}

func clientIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "0.0.0.0"
	}
	return host
}

func AuditMiddleware(pub audit.Publisher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ev := &audit.AuditEvent{
				Timestamp: start,
				Method:    r.Method,
				Path:      r.URL.Path,
				SourceIP:  clientIP(r.RemoteAddr),
				UserAgent: r.UserAgent(),
			}

			aw := &auditResponseWriter{ResponseWriter: w}

			defer func() {
				ev.DurationMs = time.Since(start).Milliseconds()
				ev.StatusCode = aw.statusCode
				pub.Publish(ev)
			}()

			ctx := context.WithValue(r.Context(), auditEventKey{}, ev)
			next.ServeHTTP(aw, r.WithContext(ctx))
		})
	}
}
