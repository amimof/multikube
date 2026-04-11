import { defineStore } from 'pinia'
import { api } from '@/api/client'
import type { V1Policy } from '@/generated/policy'

export const usePolicyStore = defineStore('policy', {
	state: () => ({
		items: [] as V1Policy[],
		loading: false,
		error: null as string | null,
	}),

	actions: {
		async fetchPolicies() {
			this.loading = true
			this.error = null

			try {
				const response = await api.policyService.policyServiceList()
				this.items = response.policys ?? []
			} catch (err) {
				this.error = err instanceof Error ? err.message : 'Failed to load policies'
				throw err
			} finally {
				this.loading = false
			}
		},

		async createPolicy(policy: V1Policy) {
			const response = await api.policyService.policyServiceCreate({ policy })
			await this.fetchPolicies()
			return response.policy
		},

		async updatePolicy(policy: V1Policy) {
			const uid = policy.meta?.uid

			if (!uid) {
				throw new Error('Policy is missing uid')
			}

			const response = await api.policyService.policyServiceUpdate({
				uid,
				name: policy.meta?.name,
				policy,
			})

			await this.fetchPolicies()
			return response.policy
		},

		async deletePolicy(policy: V1Policy) {
			const uid = policy.meta?.uid

			if (!uid) {
				throw new Error('Policy is missing uid')
			}

			await api.policyService.policyServiceDelete({
				uid,
				name: policy.meta?.name,
			})

			await this.fetchPolicies()
		},
	},
})
