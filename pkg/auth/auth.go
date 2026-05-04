package auth

import "slices"

type Permission string

const (
	BackendRead   Permission = "backend.read"
	BackendCreate Permission = "backend.create"
	BackendUpdate Permission = "backend.update"
	BackendDelete Permission = "backend.delete"

	CertificateAuthorityRead   Permission = "certificateauthority.read"
	CertificateAuthorityCreate Permission = "certificateauthority.create"
	CertificateAuthorityUpdate Permission = "certificateauthority.update"
	CertificateAuthorityDelete Permission = "certificateauthority.delete"

	CertificateRead   Permission = "certificate.read"
	CertificateCreate Permission = "certificate.create"
	CertificateUpdate Permission = "certificate.update"
	CertificateDelete Permission = "certificate.delete"

	CredentialRead   Permission = "credential.read"
	CredentialCreate Permission = "credential.create"
	CredentialUpdate Permission = "credential.update"
	CredentialDelete Permission = "credential.delete"

	PolicyRead   Permission = "policy.read"
	PolicyCreate Permission = "policy.create"
	PolicyUpdate Permission = "policy.update"
	PolicyDelete Permission = "policy.delete"

	RouteRead   Permission = "route.read"
	RouteCreate Permission = "route.create"
	RouteUpdate Permission = "route.update"
	RouteDelete Permission = "route.delete"

	UserRead   Permission = "user.read"
	UserCreate Permission = "user.create"
	UserUpdate Permission = "user.update"
	UserDelete Permission = "user.delete"

	TokenCreate Permission = "token.create"
	TokenDelete Permission = "token.delete"

	MetricsRead Permission = "metrics.read"
)

var RolePermission = map[string][]Permission{
	"admin": {
		BackendRead,
		BackendCreate,
		BackendUpdate,
		BackendDelete,
		CertificateAuthorityRead,
		CertificateAuthorityCreate,
		CertificateAuthorityUpdate,
		CertificateAuthorityDelete,
		CertificateRead,
		CertificateCreate,
		CertificateUpdate,
		CertificateDelete,
		CredentialRead,
		CredentialCreate,
		CredentialUpdate,
		CredentialDelete,
		PolicyRead,
		PolicyCreate,
		PolicyUpdate,
		PolicyDelete,
		RouteRead,
		RouteCreate,
		RouteUpdate,
		RouteDelete,
		UserRead,
		UserCreate,
		UserUpdate,
		UserDelete,
		TokenCreate,
		TokenDelete,
		MetricsRead,
	},
	"viewer": {
		BackendRead,
		CertificateAuthorityRead,
		CertificateRead,
		CredentialRead,
		PolicyRead,
		RouteRead,
		UserRead,
		MetricsRead,
	},
	"client": {},
}

var PrivateMethods = map[string]Permission{
	"/backend.v1.BackendService/Get":          BackendRead,
	"/backend.v1.BackendService/List":         BackendRead,
	"/backend.v1.BackendService/Create":       BackendCreate,
	"/backend.v1.BackendService/Update":       BackendUpdate,
	"/backend.v1.BackendService/UpdateStatus": BackendUpdate,
	"/backend.v1.BackendService/Delete":       BackendDelete,

	"/certificate_authority.v1.CertificateAuthorityService/Get":    CertificateAuthorityRead,
	"/certificate_authority.v1.CertificateAuthorityService/List":   CertificateAuthorityRead,
	"/certificate_authority.v1.CertificateAuthorityService/Create": CertificateAuthorityCreate,
	"/certificate_authority.v1.CertificateAuthorityService/Update": CertificateAuthorityUpdate,
	"/certificate_authority.v1.CertificateAuthorityService/Delete": CertificateAuthorityDelete,

	"/certificate.v1.CertificateService/Get":    CertificateRead,
	"/certificate.v1.CertificateService/List":   CertificateRead,
	"/certificate.v1.CertificateService/Create": CertificateCreate,
	"/certificate.v1.CertificateService/Update": CertificateUpdate,
	"/certificate.v1.CertificateService/Delete": CertificateDelete,

	"/credential.v1.CredentialService/Get":    CredentialRead,
	"/credential.v1.CredentialService/List":   CredentialRead,
	"/credential.v1.CredentialService/Create": CredentialCreate,
	"/credential.v1.CredentialService/Update": CredentialUpdate,
	"/credential.v1.CredentialService/Delete": CredentialDelete,

	"/policy.v1.PolicyService/Get":    PolicyRead,
	"/policy.v1.PolicyService/List":   PolicyRead,
	"/policy.v1.PolicyService/Create": PolicyCreate,
	"/policy.v1.PolicyService/Update": PolicyUpdate,
	"/policy.v1.PolicyService/Delete": PolicyDelete,

	"/route.v1.RouteService/Get":          RouteRead,
	"/route.v1.RouteService/List":         RouteRead,
	"/route.v1.RouteService/Create":       RouteCreate,
	"/route.v1.RouteService/Update":       RouteUpdate,
	"/route.v1.RouteService/UpdateStatus": RouteUpdate,
	"/route.v1.RouteService/Delete":       RouteDelete,

	"/user.v1.UserService/Get":    UserRead,
	"/user.v1.UserService/List":   UserRead,
	"/user.v1.UserService/Create": UserCreate,
	"/user.v1.UserService/Update": UserUpdate,
	"/user.v1.UserService/Delete": UserDelete,

	"/token.v1.TokenService/Issue":  TokenCreate,
	"/token.v1.TokenService/Revoke": TokenDelete,

	"/metrics.v1.MetricsService/Get": MetricsRead,
}

var PublicMethods = map[string]struct{}{
	"/auth.v1.AuthService/Login":   {},
	"/auth.v1.AuthService/Logout":  {},
	"/grpc.health.v1.Health/Check": {},
	"/grpc.health.v1.Health/List":  {},
	"/grpc.health.v1.Health/Watch": {},
}

func HasPermission(roles []string, required Permission) bool {
	for _, role := range roles {
		if slices.Contains(RolePermission[role], required) {
			return true
		}
	}
	return false
}

func IsPublicMethod(fullMethod string) bool {
	_, ok := PublicMethods[fullMethod]
	return ok
}
