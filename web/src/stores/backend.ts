import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
import type { V1Backend } from '@/generated/backend'

export const useBackendStore = defineStore('backend', {
  state: () => ({
    items: [] as V1Backend[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchBackends() {
      this.loading = true
      this.error = null

      try {
        const response = await api.backendService.backendServiceList()
        this.items = response.backends ?? []
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load backends'
        throw err
      } finally {
        this.loading = false
      }
    },

    async createBackend(backend: V1Backend) {
      const response = await api.backendService.backendServiceCreate({
        backend: sanitizePayload(backend) as V1Backend,
      })
      await this.fetchBackends()
      return response.backend
    },

    async updateBackend(backend: V1Backend) {
      const name = backend.meta?.name

      if (!name) {
        throw new Error('Backend is missing name')
      }

      const response = await api.backendService.backendServiceUpdate2({
        name,
        backend: sanitizePayload(backend) as V1Backend,
      })

      await this.fetchBackends()
      return response.backend
    },

    async deleteBackend(backend: V1Backend) {
      const name = backend.meta?.name

      if (!name) {
        throw new Error('Backend is missing name')
      }

      await api.backendService.backendServiceDelete2({
        name,
      })

      await this.fetchBackends()
    },
  },
})
