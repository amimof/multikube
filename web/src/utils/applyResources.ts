import { parseAllDocuments } from 'yaml'
import { useBackendStore } from '@/stores/backend'
import { useRouteStore } from '@/stores/route'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import { usePolicyStore } from '@/stores/policy'

const SUPPORTED_VERSIONS = [
  'backend/v1',
  'route/v1',
  'certificate_authority/v1',
  'credential/v1',
  'certificate/v1',
  'policy/v1',
] as const

type SupportedVersion = (typeof SUPPORTED_VERSIONS)[number]

export interface ApplyResult {
  index: number
  version: string
  name: string
  action: 'created' | 'updated' | 'failed'
  error?: string
}

function isSupportedVersion(v: string): v is SupportedVersion {
  return SUPPORTED_VERSIONS.includes(v as SupportedVersion)
}

/**
 * Check if an error is a 409 Conflict from the generated ResponseError class.
 * Each generated runtime.ts has its own ResponseError class, so we check
 * by error name and response status rather than instanceof.
 */
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

/**
 * Parse multi-document YAML text and validate each document.
 * Returns an array of parsed & validated documents or throws with a descriptive error.
 */
function parseDocuments(yamlText: string): { version: string; name: string; doc: Record<string, unknown> }[] {
  const parsed = parseAllDocuments(yamlText)
  const results: { version: string; name: string; doc: Record<string, unknown> }[] = []

  for (let i = 0; i < parsed.length; i++) {
    const yamlDoc = parsed[i]!

    // Check for YAML parse errors
    if (yamlDoc.errors.length > 0) {
      const errMsgs = yamlDoc.errors.map((e) => e.message).join('; ')
      throw new Error(`Document #${i + 1}: YAML parse error: ${errMsgs}`)
    }

    const doc = yamlDoc.toJSON()

    // Skip null/empty documents (e.g. trailing ---)
    if (doc == null) continue

    if (typeof doc !== 'object' || Array.isArray(doc)) {
      throw new Error(`Document #${i + 1}: must be a YAML object, got ${Array.isArray(doc) ? 'array' : typeof doc}`)
    }

    const version = doc.version
    if (!version || typeof version !== 'string') {
      throw new Error(`Document #${i + 1}: missing required "version" field`)
    }

    if (!isSupportedVersion(version)) {
      throw new Error(
        `Document #${i + 1}: unsupported version "${version}". Supported: ${SUPPORTED_VERSIONS.join(', ')}`,
      )
    }

    const name = doc.meta?.name
    if (!name || typeof name !== 'string') {
      throw new Error(`Document #${i + 1}: missing required "meta.name" field`)
    }

    results.push({ version, name, doc: doc as Record<string, unknown> })
  }

  if (results.length === 0) {
    throw new Error('No valid documents found in YAML input')
  }

  return results
}

/**
 * Apply a single resource document with upsert semantics:
 * try create first, on 409 conflict fall back to update.
 */
async function applyOne(
  version: SupportedVersion,
  doc: Record<string, unknown>,
): Promise<'created' | 'updated'> {
  const backendStore = useBackendStore()
  const routeStore = useRouteStore()
  const caStore = useCaStore()
  const credentialStore = useCredentialStore()
  const certificateStore = useCertificateStore()
  const policyStore = usePolicyStore()

  /* eslint-disable @typescript-eslint/no-explicit-any */
  const storeMap: Record<SupportedVersion, { create: (r: any) => Promise<any>; update: (r: any) => Promise<any> }> = {
    'backend/v1': { create: (r) => backendStore.createBackend(r), update: (r) => backendStore.updateBackend(r) },
    'route/v1': { create: (r) => routeStore.createRoute(r), update: (r) => routeStore.updateRoute(r) },
    'certificate_authority/v1': { create: (r) => caStore.createCa(r), update: (r) => caStore.updateCa(r) },
    'credential/v1': {
      create: (r) => credentialStore.createCredential(r),
      update: (r) => credentialStore.updateCredential(r),
    },
    'certificate/v1': {
      create: (r) => certificateStore.createCertificate(r),
      update: (r) => certificateStore.updateCertificate(r),
    },
    'policy/v1': { create: (r) => policyStore.createPolicy(r), update: (r) => policyStore.updatePolicy(r) },
  }
  /* eslint-enable @typescript-eslint/no-explicit-any */

  const { create, update } = storeMap[version]

  try {
    await create(doc)
    return 'created'
  } catch (err) {
    if (isConflictError(err)) {
      await update(doc)
      return 'updated'
    }
    throw err
  }
}

/**
 * Parse multi-document YAML and apply each resource with upsert semantics.
 * Returns structured results for each document.
 */
export async function applyResources(yamlText: string): Promise<ApplyResult[]> {
  const docs = parseDocuments(yamlText)
  const results: ApplyResult[] = []

  for (let i = 0; i < docs.length; i++) {
    const { version, name, doc } = docs[i]!
    try {
      const action = await applyOne(version as SupportedVersion, doc)
      results.push({ index: i + 1, version, name, action })
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      results.push({ index: i + 1, version, name, action: 'failed', error: errorMsg })
    }
  }

  return results
}

export { SUPPORTED_VERSIONS }
