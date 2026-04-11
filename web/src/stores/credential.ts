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

		async createCredential(credential: V1Credential) {
			const response = await api.credentialService.credentialServiceCreate({ credential })
			await this.fetchCredentials()
			return response.credential
		},

		async updateCredential(credential: V1Credential) {
			const uid = credential.meta?.uid

			if (!uid) {
				throw new Error('Credential is missing uid')
			}

			const response = await api.credentialService.credentialServiceUpdate({
				uid,
				name: credential.meta?.name,
				credential,
			})

			await this.fetchCredentials()
			return response.credential
		},

		async deleteCredential(credential: V1Credential) {
			const uid = credential.meta?.uid

			if (!uid) {
				throw new Error('Credential is missing uid')
			}

			await api.credentialService.credentialServiceDelete({
				uid,
				name: credential.meta?.name,
			})

			await this.fetchCredentials()
		},
	},
})
