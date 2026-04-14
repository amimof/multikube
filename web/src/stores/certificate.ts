import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
import type { V1Certificate } from '@/generated/certificate'

export const useCertificateStore = defineStore('certificate', {
  state: () => ({
    items: [] as V1Certificate[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
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
  },
})
