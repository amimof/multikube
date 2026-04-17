import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
import type { V1Certificate } from '@/generated/certificate'

export const useCertificateStore = defineStore('certificate', {
  state: () => ({
    items: [] as V1Certificate[],
    current: null as V1Certificate | null,
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchCertificate(name: string) {
      this.loading = true
      this.error = null

      try {
        const response = await api.certificateService.certificateServiceGet2({ name })
        this.current = response.certificate ?? null
      } catch (err) {
        this.current = null
        this.error = err instanceof Error ? err.message : 'Failed to load certificate'
        throw err
      } finally {
        this.loading = false
      }
    },

    clearCurrent() {
      this.current = null
      this.error = null
    },

    async fetchCertificates() {
      this.loading = true
      this.error = null

      try {
        const response = await api.certificateService.certificateServiceList()
        this.items = response.certificates ?? []
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load certificates'
        throw err
      } finally {
        this.loading = false
      }
    },

    async createCertificate(certificate: V1Certificate) {
      const response = await api.certificateService.certificateServiceCreate({
        certificate: sanitizePayload(certificate) as V1Certificate,
      })
      await this.fetchCertificates()
      return response.certificate
    },

    async updateCertificate(certificate: V1Certificate) {
      const name = certificate.meta?.name

      if (!name) {
        throw new Error('Certificate is missing name')
      }

      const response = await api.certificateService.certificateServiceUpdate2({
        name,
        certificate: sanitizePayload(certificate) as V1Certificate,
      })

      await this.fetchCertificates()
      return response.certificate
    },

    async deleteCertificate(certificate: V1Certificate) {
      const name = certificate.meta?.name

      if (!name) {
        throw new Error('Certificate is missing name')
      }

      await api.certificateService.certificateServiceDelete2({
        name,
      })

      await this.fetchCertificates()
    },

    async deleteManyCertificates(certificates: V1Certificate[]): Promise<{ succeeded: number; failed: { name: string; error: string }[] }> {
      const results = await Promise.allSettled(
        certificates.map(async (c) => {
          const name = c.meta?.name
          if (!name) throw new Error('Certificate is missing name')
          await api.certificateService.certificateServiceDelete2({ name })
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
            name: certificates[i]?.meta?.name ?? 'unknown',
            error: r.reason instanceof Error ? r.reason.message : 'Delete failed',
          })
        }
      })

      await this.fetchCertificates()
      return { succeeded, failed }
    },
  },
})
