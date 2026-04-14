import { BackendServiceApi, Configuration as BackendConfiguration } from '@/generated/backend'
import { CertificateAuthorityServiceApi, Configuration as CaConfiguration } from '@/generated/ca'
import { CredentialServiceApi, Configuration as CredentialConfiguration } from '@/generated/credential'
import { RouteServiceApi, Configuration as RouteConfiguration } from '@/generated/route'

const clientOptions = {
  basePath: '',
  credentials: 'include' as const,
}

export const api = {
  backendService: new BackendServiceApi(new BackendConfiguration(clientOptions)),
  certificateAuthorityService: new CertificateAuthorityServiceApi(new CaConfiguration(clientOptions)),
  credentialService: new CredentialServiceApi(new CredentialConfiguration(clientOptions)),
  routeService: new RouteServiceApi(new RouteConfiguration(clientOptions)),
}

/**
 * Strip server-managed fields from a resource payload before sending to the API.
 * The generated `V1MetaToJSONTyped` truncates `created`/`updated` to date-only
 * strings which the backend rejects. Rather than editing generated code, we
 * remove those fields (and other server-managed ones) before create/update calls.
 */
export function sanitizePayload<T extends { version?: string; meta?: { name?: string; labels?: Record<string, string> }; config?: unknown }>(
  resource: T,
): { version?: string; meta: { name?: string; labels?: Record<string, string> }; config?: unknown } {
  return {
    version: resource.version,
    meta: {
      name: resource.meta?.name,
      labels: resource.meta?.labels,
    },
    config: resource.config,
  }
}
