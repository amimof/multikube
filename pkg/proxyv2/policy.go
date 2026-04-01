package proxy

import (
	"fmt"
	"slices"
	"strings"

	policyv1 "github.com/amimof/multikube/api/policy/v1"
)

// EvalResult is the outcome of evaluating all policies against a request.
type EvalResult int

const (
	EvalAllow EvalResult = iota
	EvalDeny
)

// EvalPolicies evaluates the given policies against the principal, backend and
// k8s request using deny-first semantics:
//
//  1. If any deny rule matches = deny.
//  2. If any allow rule matches = allow.
//  3. Default = deny.
func EvalPolicies(
	policies []*policyv1.Policy,
	principal *Principal,
	backend *BackendRuntime,
	req K8sRequest,
) EvalResult {
	if len(policies) == 0 {
		return EvalDeny
	}

	var hasAllow bool

	for _, policy := range policies {
		for _, rule := range policy.GetConfig().GetRules() {
			if !ruleMatchesSubject(rule, principal) {
				continue
			}
			if !ruleMatchesCluster(rule, backend) {
				continue
			}
			if !ruleMatchesResource(rule, req) {
				continue
			}
			if !ruleMatchesAction(rule, req) {
				continue
			}

			switch rule.GetEffect() {
			case policyv1.Effect_EFFECT_DENY:
				return EvalDeny
			case policyv1.Effect_EFFECT_ALLOW:
				hasAllow = true
			}
		}
	}

	if hasAllow {
		return EvalAllow
	}
	return EvalDeny
}

// ruleMatchesSubject returns true if the principal satisfies at least one
// SubjectSelector in the rule, or if the rule has no subjects (wildcard).
func ruleMatchesSubject(rule *policyv1.Rule, p *Principal) bool {
	subjects := rule.GetSubjects()
	if len(subjects) == 0 {
		return true // no subjects = wildcard
	}

	for _, sel := range subjects {
		if subjectSelectorMatches(sel, p) {
			return true
		}
	}
	return false
}

func subjectSelectorMatches(sel *policyv1.SubjectSelector, p *Principal) bool {
	// Users
	if slices.Contains(sel.GetUsers(), p.User) {
		return true
	}
	// Groups
	for _, g := range sel.GetGroups() {
		if containsStr(p.Groups, g) {
			return true
		}
	}
	// Service accounts
	for _, sa := range sel.GetServiceAccounts() {
		if containsStr(p.ServiceAccounts, sa) {
			return true
		}
	}
	// Arbitrary claims
	for _, claim := range sel.GetClaims() {
		if v, ok := p.Claims[claim.GetName()]; ok && fmt.Sprintf("%v", v) == claim.GetValue() {
			return true
		}
	}
	return false
}

// ruleMatchesCluster returns true if the backend satisfies at least one
// ClusterSelector in the rule, or if the rule has no cluster selectors (wildcard).
func ruleMatchesCluster(rule *policyv1.Rule, backend *BackendRuntime) bool {
	clusters := rule.GetClusters()
	if len(clusters) == 0 {
		return true // no cluster selectors = wildcard
	}

	if backend == nil {
		return false
	}

	for _, sel := range clusters {
		if clusterSelectorMatches(sel, backend) {
			return true
		}
	}
	return false
}

func clusterSelectorMatches(sel *policyv1.ClusterSelector, backend *BackendRuntime) bool {
	// Match by name
	if slices.Contains(sel.GetNames(), backend.Name) {
		return true
	}
	// Match by labels
	for k, v := range sel.GetLabels() {
		if backend.Labels[k] == v {
			return true
		}
	}
	return false
}

// ruleMatchesResource returns true if the k8s request matches at least one
// ResourceSelector, or if the rule has no resource selectors (wildcard).
func ruleMatchesResource(rule *policyv1.Rule, req K8sRequest) bool {
	resources := rule.GetResources()
	if len(resources) == 0 {
		return true // no resource selectors = wildcard
	}

	for _, sel := range resources {
		if resourceSelectorMatches(sel, req) {
			return true
		}
	}
	return false
}

func resourceSelectorMatches(sel *policyv1.ResourceSelector, req K8sRequest) bool {
	if sel.GetApiGroup() != "" && sel.GetApiGroup() != req.APIGroup {
		return false
	}
	if sel.GetResource() != "" && !strings.EqualFold(sel.GetResource(), req.Resource) {
		return false
	}
	if sel.GetSubResource() != "" && !strings.EqualFold(sel.GetSubResource(), req.SubResource) {
		return false
	}
	if len(sel.GetNamespaces()) > 0 && !containsStr(sel.GetNamespaces(), req.Namespace) {
		return false
	}
	if len(sel.GetNames()) > 0 && !containsStr(sel.GetNames(), req.Name) {
		return false
	}
	return true
}

// ruleMatchesAction returns true if the HTTP verb maps to at least one allowed
// Action in the rule, or if the rule has no actions (wildcard).
func ruleMatchesAction(rule *policyv1.Rule, req K8sRequest) bool {
	actions := rule.GetActions()
	if len(actions) == 0 {
		return true // no actions = wildcard
	}

	for _, action := range actions {
		if actionMatchesVerb(action, req.Verb) {
			return true
		}
	}
	return false
}

func actionMatchesVerb(action policyv1.Action, verb string) bool {
	switch action {
	case policyv1.Action_ACTION_GET:
		return verb == "get"
	case policyv1.Action_ACTION_LIST:
		return verb == "list" || verb == "get"
	case policyv1.Action_ACTION_WATCH:
		return verb == "watch"
	case policyv1.Action_ACTION_CREATE:
		return verb == "create"
	case policyv1.Action_ACTION_UPDATE:
		return verb == "update"
	case policyv1.Action_ACTION_PATCH:
		return verb == "patch"
	case policyv1.Action_ACTION_DELETE:
		return verb == "delete"
	case policyv1.Action_ACTION_DELETECOLLECTION:
		return verb == "deletecollection"
	case policyv1.Action_ACTION_LOGS:
		return verb == "get" // logs are subresource GETs
	// ACTION_EXEC, ACTION_PORTFORWARD, ACTION_PROXY are deferred — never match.
	default:
		return false
	}
}

func containsStr(slice []string, s string) bool {
	return slices.Contains(slice, s)
}
