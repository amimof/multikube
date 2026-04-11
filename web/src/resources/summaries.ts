import type {
  BackendResource,
  CertificateAuthorityResource,
  CertificateResource,
  CredentialResource,
  PolicyResource,
  ResourceSummary,
  RouteResource,
} from '@/types/resources'

function formatText(value?: string | number | boolean | null) {
  if (value === undefined || value === null || value === '') {
    return '-'
  }

  return String(value)
}

export const backendSummary: ResourceSummary = {
  name: 'backends',
  route: '/backends',
  label: 'Backends',
  itemLabel: 'Backend',
  serviceKey: 'backendService',
  listKey: 'backends',
  responseKey: 'backend',
  endpoint: '/api/v1/backends',
  columns: [
    {
      key: 'servers',
      label: 'Servers',
      value: (item) => formatText((item as BackendResource).config?.servers?.length ?? 0),
    },
    {
      key: 'type',
      label: 'Strategy',
      value: (item) => formatText((item as BackendResource).config?.type),
    },
  ],
  createEmpty: () => ({
    version: 'backend/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      servers: [],
      insecureSkipTlsVerify: false,
    },
  }),
}

export const certificateAuthoritySummary: ResourceSummary = {
  name: 'cas',
  route: '/cas',
  label: "CA's",
  itemLabel: 'CA',
  serviceKey: 'certificateAuthorityService',
  listKey: 'certificateAuthoritys',
  responseKey: 'certificateAuthority',
  endpoint: '/api/v1/certificate_authoritys',
  columns: [
    {
      key: 'certificate',
      label: 'Certificate Source',
      value: (item) => {
        const resource = item as CertificateAuthorityResource
        return formatText(resource.config?.certificate || (resource.config?.certificateData ? 'inline data' : ''))
      },
    },
  ],
  createEmpty: () => ({
    version: 'certificate_authority/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      certificateData: '',
    },
  }),
}

export const certificateSummary: ResourceSummary = {
  name: 'certificates',
  route: '/certificates',
  label: 'Certificates',
  itemLabel: 'Certificate',
  serviceKey: 'certificateService',
  listKey: 'certificates',
  responseKey: 'certificate',
  endpoint: '/api/v1/certificates',
  columns: [
    {
      key: 'certificate',
      label: 'Certificate Source',
      value: (item) => {
        const resource = item as CertificateResource
        return formatText(resource.config?.certificate || (resource.config?.certificateData ? 'inline data' : ''))
      },
    },
    {
      key: 'key',
      label: 'Key Source',
      value: (item) => {
        const resource = item as CertificateResource
        return formatText(resource.config?.key || (resource.config?.keyData ? 'inline data' : ''))
      },
    },
  ],
  createEmpty: () => ({
    version: 'certificate/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      certificateData: '',
      keyData: '',
    },
  }),
}

export const credentialSummary: ResourceSummary = {
  name: 'credentials',
  route: '/credentials',
  label: 'Credentials',
  itemLabel: 'Credential',
  serviceKey: 'credentialService',
  listKey: 'credentials',
  responseKey: 'credential',
  endpoint: '/api/v1/credentials',
  columns: [
    {
      key: 'type',
      label: 'Type',
      value: (item) => {
        const resource = item as CredentialResource

        if (resource.config?.basic) {
          return 'basic'
        }

        if (resource.config?.token) {
          return 'token'
        }

        if (resource.config?.clientCertificateRef) {
          return 'client certificate'
        }

        return '-'
      },
    },
    {
      key: 'healthy',
      label: 'Healthy',
      value: (item) => formatText((item as CredentialResource).status?.healthy),
    },
  ],
  createEmpty: () => ({
    version: 'credential/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      token: '',
    },
  }),
}

export const policySummary: ResourceSummary = {
  name: 'policies',
  route: '/policies',
  label: 'Policies',
  itemLabel: 'Policy',
  serviceKey: 'policyService',
  listKey: 'policys',
  responseKey: 'policy',
  endpoint: '/api/v1/policys',
  columns: [
    {
      key: 'rules',
      label: 'Rules',
      value: (item) => formatText((item as PolicyResource).config?.rules?.length ?? 0),
    },
  ],
  createEmpty: () => ({
    version: 'policy/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      rules: [],
    },
  }),
}

export const routeSummary: ResourceSummary = {
  name: 'routes',
  route: '/routes',
  label: 'Routes',
  itemLabel: 'Route',
  serviceKey: 'routeService',
  listKey: 'routes',
  responseKey: 'route',
  endpoint: '/api/v1/routes',
  columns: [
    {
      key: 'backendRef',
      label: 'Backend Ref',
      value: (item) => formatText((item as RouteResource).config?.backendRef),
    },
    {
      key: 'match',
      label: 'Match',
      value: (item) => {
        const match = (item as RouteResource).config?.match

        if (!match) {
          return '-'
        }

        if (match.sni) {
          return `sni:${match.sni}`
        }

        if (match.path) {
          return `path:${match.path}`
        }

        if (match.pathPrefix) {
          return `prefix:${match.pathPrefix}`
        }

        if (match.header?.name) {
          return `header:${match.header.name}`
        }

        if (match.jwt?.claim) {
          return `jwt:${match.jwt.claim}`
        }

        return 'custom'
      },
    },
  ],
  createEmpty: () => ({
    version: 'route/v1',
    meta: { name: '', labels: {} },
    config: {
      name: '',
      backendRef: '',
      match: {
        sni: '',
      },
    },
  }),
}
