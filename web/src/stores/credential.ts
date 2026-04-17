import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
import type { V1Credential } from '@/generated/credential'

export const useCredentialStore = defineStore('credential', {
  state: () => ({
    items: [] as V1Credential[],
    current: null as V1Credential | null,
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchCredential(name: string) {
      this.loading = true
      this.error = null

      try {
        const response = await api.credentialService.credentialServiceGet2({ name })
        this.current = response.credential ?? null
      } catch (err) {
        this.current = null
        this.error = err instanceof Error ? err.message : 'Failed to load credential'
        throw err
      } finally {
        this.loading = false
      }
    },

    clearCurrent() {
      this.current = null
      this.error = null
    },

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

    async createCredential(credential: V1Credential) {
      const response = await api.credentialService.credentialServiceCreate({
        credential: sanitizePayload(credential) as V1Credential,
      })
      await this.fetchCredentials()
      return response.credential
    },

    async updateCredential(credential: V1Credential) {
      const name = credential.meta?.name

      if (!name) {
        throw new Error('Credential is missing name')
      }

      const response = await api.credentialService.credentialServiceUpdate2({
        name,
        credential: sanitizePayload(credential) as V1Credential,
      })

      await this.fetchCredentials()
      return response.credential
    },

    async deleteCredential(credential: V1Credential) {
      const name = credential.meta?.name

      if (!name) {
        throw new Error('Credential is missing name')
      }

      await api.credentialService.credentialServiceDelete2({
        name,
      })

      await this.fetchCredentials()
    },

    async deleteManyCredentials(credentials: V1Credential[]): Promise<{ succeeded: number; failed: { name: string; error: string }[] }> {
      const results = await Promise.allSettled(
        credentials.map(async (c) => {
          const name = c.meta?.name
          if (!name) throw new Error('Credential is missing name')
          await api.credentialService.credentialServiceDelete2({ name })
          return name
        }),
      )

      const failed: { name: string; error: string }[] = []
      let succeeded = 0

      results.forEach((r, i) => {
        if (r.status === 'fulfilled') {
          succeeded++
        } else {
          failed.push({
            name: credentials[i]?.meta?.name ?? 'unknown',
            error: r.reason instanceof Error ? r.reason.message : 'Delete failed',
          })
        }
      })

      await this.fetchCredentials()
      return { succeeded, failed }
    },
  },
})
