import type { V1Backend, V1Probe } from '@/generated/backend'

export function defaultHealthProbe(): V1Probe {
  return {
    path: '/healthz',
    timeoutSeconds: '1',
    periodSeconds: '5',
    failureThreshold: '3',
    successThreshold: '3',
    initialDelaySeconds: '1',
  }
}

export function defaultReadyProbe(): V1Probe {
  return {
    path: '/readyz',
    timeoutSeconds: '1',
    periodSeconds: '5',
    failureThreshold: '3',
    successThreshold: '3',
    initialDelaySeconds: '1',
  }
}

export function normalizeBackendForm(resource: V1Backend): V1Backend {
  const normalized = structuredClone(resource)

  if (!normalized.config) normalized.config = {}

  if (!normalized.config.impersonationConfig) {
    normalized.config.impersonationConfig = {
      name: 'default',
      enabled: true,
      usernameClaim: 'sub',
      groupsClaim: 'groups',
      extraClaims: [],
    }
  }

  if (!normalized.config.probes) normalized.config.probes = {}

  normalized.config.probes.healthiness = {
    ...defaultHealthProbe(),
    ...normalized.config.probes.healthiness,
  }
  normalized.config.probes.readiness = {
    ...defaultReadyProbe(),
    ...normalized.config.probes.readiness,
  }

  return normalized
}
