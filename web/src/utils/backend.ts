import type { V1TargetStatus } from '@/generated/backend'

export const lbTypeLabels: Record<string, string> = {
  LOAD_BALANCING_TYPE_UNSPECIFIED: 'Unspecified',
  LOAD_BALANCING_TYPE_ROUND_ROBIN: 'Round Robin',
  LOAD_BALANCING_TYPE_LEAST_CONNECTIONS: 'Least Connections',
  LOAD_BALANCING_TYPE_RANDOM: 'Random',
  LOAD_BALANCING_TYPE_WEIGHTED_ROUND_ROBIN: 'Weighted Round Robin',
}

export function lbLabel(type?: string): string {
  if (!type) return '-'
  return lbTypeLabels[type] ?? type
}

/**
 * Total is always based on configured servers, not observed statuses.
 */
export function countReadyServers(
  servers: string[],
  targetStatuses?: Record<string, V1TargetStatus> | undefined,
): number {
  if (!targetStatuses) return 0
  return servers.filter((url) => targetStatuses[url]?.readiness?.isReady === true).length
}

export function countHealthyServers(
  servers: string[],
  targetStatuses?: Record<string, V1TargetStatus> | undefined,
): number {
  if (!targetStatuses) return 0
  return servers.filter((url) => targetStatuses[url]?.healthiness?.isHealthy === true).length
}

export function countTotalServers(servers: string[]): number {
  return servers.length
}

export type HealthTagType = 'success' | 'danger' | 'warning' | 'info'

export function healthTagType(healthy: number, total: number): HealthTagType {
  if (total === 0) return 'info'
  if (healthy === total) return 'success'
  if (healthy === 0) return 'danger'
  return 'warning'
}

export function readinessLabel(isReady?: boolean): string {
  if (isReady === true) return 'Ready'
  if (isReady === false) return 'Not Ready'
  return 'Unknown'
}

export function healthinessLabel(isHealthy?: boolean): string {
  if (isHealthy === true) return 'Healthy'
  if (isHealthy === false) return 'Unhealthy'
  return 'Unknown'
}

export function booleanStatusTagType(value?: boolean): HealthTagType {
  if (value === true) return 'success'
  if (value === false) return 'danger'
  return 'info'
}

export type TargetVisualState = 'healthy' | 'degraded' | 'unknown'

export function targetVisualState(status?: V1TargetStatus): TargetVisualState {
  if (!status) return 'unknown'
  if (status.readiness?.isReady === false || status.healthiness?.isHealthy === false) return 'degraded'
  if (status.readiness?.isReady === true && status.healthiness?.isHealthy === true) return 'healthy'
  return 'unknown'
}
