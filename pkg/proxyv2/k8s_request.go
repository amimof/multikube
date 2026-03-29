package proxy

import (
	"net/http"
	"strings"
)

// K8sRequest holds the parsed components of a Kubernetes API request path.
// It is used during policy evaluation to match ResourceSelector rules.
type K8sRequest struct {
	APIGroup    string
	Resource    string
	SubResource string
	Namespace   string
	Name        string
	Verb        string
}

// ParseK8sRequest extracts Kubernetes API attributes from an HTTP request.
// It handles both core API paths (/api/v1/...) and named group paths
// (/apis/<group>/<version>/...).
//
// Supported path shapes:
//
//	/api/v1/namespaces/<ns>/<resource>/<name>/<subresource>
//	/apis/<group>/<version>/namespaces/<ns>/<resource>/<name>/<subresource>
//	/api/v1/<resource>/<name>
//	/apis/<group>/<version>/<resource>/<name>
func ParseK8sRequest(r *http.Request) K8sRequest {
	verb := httpMethodToVerb(r.Method, r.URL.Path)

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")

	req := K8sRequest{Verb: verb}

	switch {
	case len(parts) >= 2 && parts[0] == "api":
		// Core group: /api/v1/...
		// parts[0]=api parts[1]=v1 parts[2..]=rest
		req.APIGroup = ""
		parseResourcePath(&req, parts[2:])

	case len(parts) >= 3 && parts[0] == "apis":
		// Named group: /apis/<group>/<version>/...
		// parts[0]=apis parts[1]=group parts[2]=version parts[3..]=rest
		req.APIGroup = parts[1]
		if len(parts) > 3 {
			parseResourcePath(&req, parts[3:])
		}
	}

	return req
}

// parseResourcePath fills resource, namespace, name and sub-resource from the
// trailing segments of a Kubernetes API path (after the version prefix).
//
//	[]                               → nothing
//	[resource]                       → cluster-scoped list
//	[resource, name]                 → cluster-scoped get
//	[resource, name, subresource]    → cluster-scoped subresource
//	[namespaces, ns, resource]       → namespaced list
//	[namespaces, ns, resource, name] → namespaced get
//	[namespaces, ns, resource, name, subresource] → namespaced subresource
func parseResourcePath(req *K8sRequest, rest []string) {
	if len(rest) == 0 {
		return
	}

	if rest[0] == "namespaces" && len(rest) >= 3 {
		req.Namespace = safeGet(rest, 1)
		req.Resource = safeGet(rest, 2)
		req.Name = safeGet(rest, 3)
		req.SubResource = safeGet(rest, 4)
		return
	}

	req.Resource = safeGet(rest, 0)
	req.Name = safeGet(rest, 1)
	req.SubResource = safeGet(rest, 2)
}

func safeGet(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}

// httpMethodToVerb maps an HTTP method to a Kubernetes verb.
func httpMethodToVerb(method, path string) string {
	switch strings.ToUpper(method) {
	case http.MethodGet:
		// Distinguish list (no name in last segment) from get — a best-effort
		// heuristic; policy eval will check both verbs anyway.
		if strings.HasSuffix(path, "/watch") {
			return "watch"
		}
		return "get"
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodPatch:
		return "patch"
	case http.MethodDelete:
		return "delete"
	default:
		return strings.ToLower(method)
	}
}
