package audit

import (
	"sync/atomic"
	"time"
)

type OverflowPolicy string

const (
	OverflowDrop  OverflowPolicy = "drop"
	OverflowBlock OverflowPolicy = "block"
)

type AuditEvent struct {
	Timestamp time.Time `json:"ts"`
	RequestID string    `json:"request_id,omitempty"`
	Subject   string    `json:"subject,omitempty"`
	Username  string    `json:"username,omitempty"`
	Groups    []string  `json:"groups,omitempty"`
	Issuer    string    `json:"issuer,omitempty"`

	Cluster string `json:"cluster,omitempty"`
	Backend string `json:"backend,omitempty"`
	Route   string `json:"route,omitempty"`

	Method    string `json:"method,omitempty"`
	Path      string `json:"path,omitempty"`
	SourceIP  string `json:"source_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	K8sVerb     string `json:"k8s_verb,omitempty"`
	APIGroup    string `json:"api_group,omitempty"`
	Resource    string `json:"resource,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Name        string `json:"name,omitempty"`
	Subresource string `json:"subresource,omitempty"`

	Allowed    bool   `json:"allowed"`
	StatusCode int    `json:"status_code,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`

	PolicyIDs     []string `json:"policy_ids,omitempty"`
	ConfigVersion uint64   `json:"config_version,omitempty"`
}

type Stats struct {
	Published atomic.Uint64
	Written   atomic.Uint64
	Dropped   atomic.Uint64
	Failed    atomic.Uint64
	Flushes   atomic.Uint64
}
