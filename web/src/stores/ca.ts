import { defineStore } from 'pinia'
import { api } from '@/api/client'
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
			const response = await api.certificateAuthorityService.certificateAuthorityServiceCreate({ certificateAuthority })
			await this.fetchCas()
			return response.certificateAuthority
		},

		async updateCa(certificateAuthority: V1CertificateAuthority) {
			const uid = certificateAuthority.meta?.uid

			if (!uid) {
				throw new Error('CA is missing uid')
			}

			const response = await api.certificateAuthorityService.certificateAuthorityServiceUpdate({
				uid,
				name: certificateAuthority.meta?.name,
				certificateAuthority,
			})

			await this.fetchCas()
			return response.certificateAuthority
		},

		async deleteCa(certificateAuthority: V1CertificateAuthority) {
			const uid = certificateAuthority.meta?.uid

			if (!uid) {
				throw new Error('CA is missing uid')
			}

			await api.certificateAuthorityService.certificateAuthorityServiceDelete({
				uid,
				name: certificateAuthority.meta?.name,
			})

			await this.fetchCas()
		},
	},
})
