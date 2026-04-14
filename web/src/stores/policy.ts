import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
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
      const response = await api.policyService.policyServiceCreate({
        policy: sanitizePayload(policy) as V1Policy,
      })
      await this.fetchPolicies()
      return response.policy
    },

    async updatePolicy(policy: V1Policy) {
      const name = policy.meta?.name

      if (!name) {
        throw new Error('Policy is missing name')
      }

      const response = await api.policyService.policyServiceUpdate2({
        name,
        policy: sanitizePayload(policy) as V1Policy,
      })

      await this.fetchPolicies()
      return response.policy
    },

    async deletePolicy(policy: V1Policy) {
      const name = policy.meta?.name

      if (!name) {
        throw new Error('Policy is missing name')
      }

      await api.policyService.policyServiceDelete2({
        name,
      })

      await this.fetchPolicies()
    },
  },
})
