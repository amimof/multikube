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
 * Count healthy servers from the configured server list.
 * Total is always based on configured servers, not observed statuses.
 */
export function countHealthyServers(
  servers: string[],
  targetStatuses?: Record<string, V1TargetStatus> | undefined,
): number {
  if (!targetStatuses) return 0
  return servers.filter((url) => targetStatuses[url]?.phase === 'Healthy').length
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
