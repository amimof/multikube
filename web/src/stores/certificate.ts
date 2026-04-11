import { defineStore } from 'pinia'
import { api } from '@/api/client'
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
			const response = await api.certificateService.certificateServiceCreate({ certificate })
			await this.fetchCertificates()
			return response.certificate
		},

		async updateCertificate(certificate: V1Certificate) {
			const uid = certificate.meta?.uid

			if (!uid) {
				throw new Error('Certificate is missing uid')
			}

			const response = await api.certificateService.certificateServiceUpdate({
				uid,
				name: certificate.meta?.name,
				certificate,
			})

			await this.fetchCertificates()
			return response.certificate
		},

		async deleteCertificate(certificate: V1Certificate) {
			const uid = certificate.meta?.uid

			if (!uid) {
				throw new Error('Certificate is missing uid')
			}

			await api.certificateService.certificateServiceDelete({
				uid,
				name: certificate.meta?.name,
			})

			await this.fetchCertificates()
		},
	},
})
