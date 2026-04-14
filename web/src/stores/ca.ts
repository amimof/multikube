import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
import type { V1CertificateAuthority } from '@/generated/ca'

export const useCaStore = defineStore('ca', {
  state: () => ({
    items: [] as V1CertificateAuthority[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchCas() {
      this.loading = true
      this.error = null

      try {
        const response = await api.certificateAuthorityService.certificateAuthorityServiceList()
        this.items = response.certificateAuthoritys ?? []
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load CAs'
        throw err
      } finally {
        this.loading = false
      }
    },

    async createCa(certificateAuthority: V1CertificateAuthority) {
      const response = await api.certificateAuthorityService.certificateAuthorityServiceCreate({
        certificateAuthority: sanitizePayload(certificateAuthority) as V1CertificateAuthority,
      })
      await this.fetchCas()
      return response.certificateAuthority
    },

    async updateCa(certificateAuthority: V1CertificateAuthority) {
      const name = certificateAuthority.meta?.name

      if (!name) {
        throw new Error('CA is missing name')
      }

      const response = await api.certificateAuthorityService.certificateAuthorityServiceUpdate2({
        name,
        certificateAuthority: sanitizePayload(certificateAuthority) as V1CertificateAuthority,
      })

      await this.fetchCas()
      return response.certificateAuthority
    },

    async deleteCa(certificateAuthority: V1CertificateAuthority) {
      const name = certificateAuthority.meta?.name

      if (!name) {
        throw new Error('CA is missing name')
      }

      await api.certificateAuthorityService.certificateAuthorityServiceDelete2({
        name,
      })

      await this.fetchCas()
    },

    async deleteManyCas(cas: V1CertificateAuthority[]): Promise<{ succeeded: number; failed: { name: string; error: string }[] }> {
      const results = await Promise.allSettled(
        cas.map(async (ca) => {
          const name = ca.meta?.name
          if (!name) throw new Error('CA is missing name')
          await api.certificateAuthorityService.certificateAuthorityServiceDelete2({ name })
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
            name: cas[i]?.meta?.name ?? 'unknown',
            error: r.reason instanceof Error ? r.reason.message : 'Delete failed',
          })
        }
      })

      await this.fetchCas()
      return { succeeded, failed }
    },
  },
})
