package audit

import (
	"context"

	auditv1 "github.com/amimof/multikube/api/audit/v1"
)

type Subscriber interface {
	Subscribe(ctx context.Context, fromSeq uint64, live bool) (<-chan *auditv1.AuditEntry, error)
}
