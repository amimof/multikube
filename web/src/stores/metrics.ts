import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { V1GetResponse } from '@/generated/metrics'

export const useMetricsStore = defineStore('metrics', {
  state: () => ({
    data: null as V1GetResponse | null,
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchMetrics() {
      this.loading = true
      this.error = null

      try {
        this.data = await api.metricsService.metricsServiceGet()
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load metrics'
        throw err
      } finally {
        this.loading = false
      }
    },
  },
})
