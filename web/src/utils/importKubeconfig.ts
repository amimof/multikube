import { parse } from 'yaml'
import { useBackendStore } from '@/stores/backend'
import { useRouteStore } from '@/stores/route'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import { usePolicyStore } from '@/stores/policy'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/**
 * Decode a base64-encoded string to UTF-8 text.
 * Kubeconfig *-data fields (certificate-authority-data, client-certificate-data,
 * client-key-data) are base64-encoded per the kubeconfig spec. The backend
 * expects raw PEM, so we must decode before importing.
 */
function decodeBase64(encoded: string): string {
  try {
    return atob(encoded)
  } catch {
    // If it fails, the value may already be raw PEM — return as-is
    return encoded
  }
}

// ---------------------------------------------------------------------------
// Kubeconfig types (subset we care about)
// ---------------------------------------------------------------------------

interface KubeconfigCluster {
  server?: string
  'certificate-authority'?: string
  'certificate-authority-data'?: string
  'insecure-skip-tls-verify'?: boolean
}

interface KubeconfigUser {
  token?: string
  'token-file'?: string
  username?: string
  password?: string
  'client-certificate'?: string
  'client-certificate-data'?: string
  'client-key'?: string
  'client-key-data'?: string
  'auth-provider'?: unknown
  exec?: unknown
}

interface KubeconfigContext {
  cluster?: string
  user?: string
  namespace?: string
}

interface Kubeconfig {
  apiVersion?: string
  kind?: string
  'current-context'?: string
  clusters?: { name: string; cluster: KubeconfigCluster }[]
  users?: { name: string; user: KubeconfigUser }[]
  contexts?: { name: string; context: KubeconfigContext }[]
}

// ---------------------------------------------------------------------------
// Import plan types
// ---------------------------------------------------------------------------

export interface ImportResourceNames {
  backend: string
  credential: string
  certificate: string
  certificateAuthority: string
}

export interface ImportPlanResource {
  kind: string
  name: string
  /** The full resource payload ready to send to the store */
  payload: Record<string, unknown>
}

export interface ImportPlan {
  contextName: string
  names: ImportResourceNames
  authType: 'none' | 'token' | 'basic' | 'client-certificate'
  resources: ImportPlanResource[]
}

export interface ImportResult {
  kind: string
  name: string
  action: 'created' | 'updated' | 'skipped' | 'failed'
  error?: string
}

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

export function parseKubeconfig(text: string): Kubeconfig {
  const doc = parse(text)
  if (doc == null || typeof doc !== 'object') {
    throw new Error('Invalid kubeconfig: not a YAML object')
  }
  return doc as Kubeconfig
}

export function getContextNames(kc: Kubeconfig): string[] {
  return (kc.contexts ?? []).map((c) => c.name).filter(Boolean)
}

export function getDefaultContext(kc: Kubeconfig): string | undefined {
  return kc['current-context'] || undefined
}

// ---------------------------------------------------------------------------
// Default resource names (mirrors CLI defaultImportResourceNames)
// ---------------------------------------------------------------------------

export function defaultImportResourceNames(contextName: string): ImportResourceNames {
  return {
    backend: `${contextName}-backend`,
    credential: `${contextName}-credential`,
    certificate: `${contextName}-certificate`,
    certificateAuthority: `${contextName}-certificate-authority`,
  }
}

// ---------------------------------------------------------------------------
// Build import plan
// ---------------------------------------------------------------------------

export function buildImportPlan(
  kc: Kubeconfig,
  contextName: string,
  nameOverrides: Partial<ImportResourceNames> = {},
): ImportPlan {
  const ctxEntry = (kc.contexts ?? []).find((c) => c.name === contextName)
  if (!ctxEntry) {
    throw new Error(`Context "${contextName}" not found in kubeconfig`)
  }
  const ctx = ctxEntry.context
  if (!ctx?.cluster) {
    throw new Error(`Context "${contextName}" does not reference a cluster`)
  }

  // Resolve cluster
  const clusterEntry = (kc.clusters ?? []).find((c) => c.name === ctx.cluster)
  if (!clusterEntry) {
    throw new Error(`Cluster "${ctx.cluster}" referenced by context "${contextName}" not found in kubeconfig`)
  }
  const cluster = clusterEntry.cluster

  // Resolve user (optional)
  let user: KubeconfigUser | undefined
  if (ctx.user) {
    const userEntry = (kc.users ?? []).find((u) => u.name === ctx.user)
    if (!userEntry) {
      throw new Error(`User "${ctx.user}" referenced by context "${contextName}" not found in kubeconfig`)
    }
    user = userEntry.user
  }

  // Merge names
  const defaults = defaultImportResourceNames(contextName)
  const names: ImportResourceNames = {
    backend: nameOverrides.backend || defaults.backend,
    credential: nameOverrides.credential || defaults.credential,
    certificate: nameOverrides.certificate || defaults.certificate,
    certificateAuthority: nameOverrides.certificateAuthority || defaults.certificateAuthority,
  }

  const resources: ImportPlanResource[] = []

  // --- Validate unsupported external file references ---
  if (cluster['certificate-authority']) {
    throw new Error(
      `Cluster "${ctx.cluster}" uses an external certificate-authority file reference. ` +
        'Browser import only supports inline data (certificate-authority-data). ' +
        'Use "multikubectl import" for file-based kubeconfigs.',
    )
  }

  // --- CA ---
  const caData = cluster['certificate-authority-data']
  if (caData) {
    resources.push({
      kind: 'Certificate Authority',
      name: names.certificateAuthority,
      payload: {
        version: 'certificate_authority/v1',
        meta: { name: names.certificateAuthority },
        config: {
          certificateData: decodeBase64(caData),
        },
      },
    })
  }

  // --- Auth validation ---
  let authType: ImportPlan['authType'] = 'none'
  if (user) {
    // Reject unsupported auth methods
    if (user['auth-provider']) {
      throw new Error('Unsupported auth method: auth-provider. Use "multikubectl import" instead.')
    }
    if (user.exec) {
      throw new Error('Unsupported auth method: exec. Use "multikubectl import" instead.')
    }

    // Reject external file references
    if (user['token-file']) {
      throw new Error(
        'User references an external token-file. ' +
          'Browser import only supports inline token. ' +
          'Use "multikubectl import" for file-based kubeconfigs.',
      )
    }
    if (user['client-certificate']) {
      throw new Error(
        'User references an external client-certificate file. ' +
          'Browser import only supports inline data (client-certificate-data). ' +
          'Use "multikubectl import" for file-based kubeconfigs.',
      )
    }
    if (user['client-key']) {
      throw new Error(
        'User references an external client-key file. ' +
          'Browser import only supports inline data (client-key-data). ' +
          'Use "multikubectl import" for file-based kubeconfigs.',
      )
    }

    const hasToken = !!user.token
    const hasBasic = !!user.username || !!user.password
    const hasClientCert = !!user['client-certificate-data'] || !!user['client-key-data']

    if (hasBasic && (!user.username || !user.password)) {
      throw new Error('Basic auth requires both username and password')
    }
    if (hasClientCert) {
      const hasCert = !!user['client-certificate-data']
      const hasKey = !!user['client-key-data']
      if (hasCert !== hasKey) {
        throw new Error('Client certificate auth requires both certificate and key data')
      }
    }

    // Count methods
    let methodCount = 0
    if (hasToken) methodCount++
    if (hasBasic) methodCount++
    if (hasClientCert) methodCount++

    if (methodCount > 1) {
      throw new Error('Kubeconfig user contains multiple supported auth methods; choose one')
    }

    // --- Certificate (for mTLS) ---
    if (hasClientCert) {
      authType = 'client-certificate'
      resources.push({
        kind: 'Certificate',
        name: names.certificate,
        payload: {
          version: 'certificate/v1',
          meta: { name: names.certificate },
          config: {
            certificateData: decodeBase64(user['client-certificate-data']!),
            keyData: decodeBase64(user['client-key-data']!),
          },
        },
      })
    }

    // --- Credential ---
    if (hasToken) {
      authType = 'token'
      resources.push({
        kind: 'Credential',
        name: names.credential,
        payload: {
          version: 'credential/v1',
          meta: { name: names.credential },
          config: {
            token: user.token,
          },
        },
      })
    } else if (hasBasic) {
      authType = 'basic'
      resources.push({
        kind: 'Credential',
        name: names.credential,
        payload: {
          version: 'credential/v1',
          meta: { name: names.credential },
          config: {
            basic: {
              username: user.username,
              password: user.password,
            },
          },
        },
      })
    } else if (hasClientCert) {
      // Credential with clientCertificateRef
      resources.push({
        kind: 'Credential',
        name: names.credential,
        payload: {
          version: 'credential/v1',
          meta: { name: names.credential },
          config: {
            clientCertificateRef: names.certificate,
          },
        },
      })
    }
  }

  // --- Backend (always created) ---
  const backendConfig: Record<string, unknown> = {
    servers: [cluster.server],
    insecureSkipTlsVerify: cluster['insecure-skip-tls-verify'] ?? false,
  }
  if (caData) {
    backendConfig.caRef = names.certificateAuthority
  }
  if (authType !== 'none') {
    backendConfig.authRef = names.credential
  }
  resources.push({
    kind: 'Backend',
    name: names.backend,
    payload: {
      version: 'backend/v1',
      meta: { name: names.backend },
      config: backendConfig,
    },
  })

  return { contextName, names, authType, resources }
}

// ---------------------------------------------------------------------------
// Execute import plan
// ---------------------------------------------------------------------------

function isConflictError(err: unknown): boolean {
  if (
    err &&
    typeof err === 'object' &&
    'name' in err &&
    (err as { name: string }).name === 'ResponseError' &&
    'response' in err
  ) {
    const response = (err as { response: Response }).response
    return response?.status === 409
  }
  return false
}

/* eslint-disable @typescript-eslint/no-explicit-any */
type StoreAction = {
  create: (r: any) => Promise<any>
  update: (r: any) => Promise<any>
}

function getStoreActions(version: string): StoreAction {
  switch (version) {
    case 'backend/v1': {
      const s = useBackendStore()
      return { create: (r) => s.createBackend(r), update: (r) => s.updateBackend(r) }
    }
    case 'certificate_authority/v1': {
      const s = useCaStore()
      return { create: (r) => s.createCa(r), update: (r) => s.updateCa(r) }
    }
    case 'credential/v1': {
      const s = useCredentialStore()
      return { create: (r) => s.createCredential(r), update: (r) => s.updateCredential(r) }
    }
    case 'certificate/v1': {
      const s = useCertificateStore()
      return { create: (r) => s.createCertificate(r), update: (r) => s.updateCertificate(r) }
    }
    case 'route/v1': {
      const s = useRouteStore()
      return { create: (r) => s.createRoute(r), update: (r) => s.updateRoute(r) }
    }
    case 'policy/v1': {
      const s = usePolicyStore()
      return { create: (r) => s.createPolicy(r), update: (r) => s.updatePolicy(r) }
    }
    default:
      throw new Error(`Unknown resource version: ${version}`)
  }
}
/* eslint-enable @typescript-eslint/no-explicit-any */

export async function executeImportPlan(plan: ImportPlan, force: boolean): Promise<ImportResult[]> {
  const results: ImportResult[] = []

  for (const resource of plan.resources) {
    const version = resource.payload.version as string
    const { create, update } = getStoreActions(version)

    try {
      await create(resource.payload)
      results.push({ kind: resource.kind, name: resource.name, action: 'created' })
    } catch (err) {
      if (isConflictError(err)) {
        if (force) {
          try {
            await update(resource.payload)
            results.push({ kind: resource.kind, name: resource.name, action: 'updated' })
          } catch (updateErr) {
            const msg = updateErr instanceof Error ? updateErr.message : String(updateErr)
            results.push({ kind: resource.kind, name: resource.name, action: 'failed', error: msg })
            // Stop on failure — downstream resources may depend on this one
            break
          }
        } else {
          results.push({
            kind: resource.kind,
            name: resource.name,
            action: 'failed',
            error: `Already exists. Enable "Force update" to overwrite.`,
          })
          // Stop — downstream resources reference this one
          break
        }
      } else {
        const msg = err instanceof Error ? err.message : String(err)
        results.push({ kind: resource.kind, name: resource.name, action: 'failed', error: msg })
        break
      }
    }
  }

  return results
}
