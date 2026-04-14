import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { V1Credential } from '@/generated/credential'

export const useCredentialStore = defineStore('credential', {
  state: () => ({
    items: [] as V1Credential[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchCredentials() {
      this.loading = true
      this.error = null

      try {
        const response = await api.credentialService.credentialServiceList()
        this.items = response.credentials ?? []
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load credentials'
        throw err
      } finally {
        this.loading = false
      }
    },
  },
})
