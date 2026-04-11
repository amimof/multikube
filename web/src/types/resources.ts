import type { V1Backend } from '@/generated/backend'
import type { V1CertificateAuthority } from '@/generated/ca'
import type { V1Certificate } from '@/generated/certificate'
import type { V1Credential } from '@/generated/credential'
import type { V1Policy } from '@/generated/policy'
import type { V1Route } from '@/generated/route'

export type BackendResource = V1Backend
export type CertificateAuthorityResource = V1CertificateAuthority
export type CertificateResource = V1Certificate
export type CredentialResource = V1Credential
export type PolicyResource = V1Policy
export type RouteResource = V1Route

export type ResourceMeta =
  | BackendResource['meta']
  | CertificateAuthorityResource['meta']
  | CertificateResource['meta']
  | CredentialResource['meta']
  | PolicyResource['meta']
  | RouteResource['meta']

export type BackendConfig = V1Backend['config']
export type BackendStatus = V1Backend['status']
export type CertificateAuthorityConfig = V1CertificateAuthority['config']
export type CertificateConfig = V1Certificate['config']
export type CredentialConfig = V1Credential['config']
export type CredentialBasic = NonNullable<NonNullable<CredentialConfig>['basic']>
export type CredentialStatus = V1Credential['status']
export type PolicyConfig = V1Policy['config']
export type RouteConfig = V1Route['config']
export type RouteMatch = NonNullable<NonNullable<RouteConfig>['match']>
export type HeaderMatch = NonNullable<RouteMatch['header']>
export type JwtMatch = NonNullable<RouteMatch['jwt']>
export type RouteStatus = V1Route['status']

export type ResourceItem =
  | BackendResource
  | CertificateAuthorityResource
  | CertificateResource
  | CredentialResource
  | PolicyResource
  | RouteResource

export interface ResourceColumn {
  key: string
  label: string
  value: (item: ResourceItem) => string
}

export interface ResourceSummary {
  name: string
  route: string
  label: string
  itemLabel: string
  serviceKey:
    | 'backendService'
    | 'certificateAuthorityService'
    | 'certificateService'
    | 'credentialService'
    | 'policyService'
    | 'routeService'
  listKey: string
  responseKey: string
  endpoint: string
  columns: ResourceColumn[]
  createEmpty: () => ResourceItem
}
