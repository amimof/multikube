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

    async deleteManyPolicies(policies: V1Policy[]): Promise<{ succeeded: number; failed: { name: string; error: string }[] }> {
      const results = await Promise.allSettled(
        policies.map(async (p) => {
          const name = p.meta?.name
          if (!name) throw new Error('Policy is missing name')
          await api.policyService.policyServiceDelete2({ name })
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
            name: policies[i]?.meta?.name ?? 'unknown',
            error: r.reason instanceof Error ? r.reason.message : 'Delete failed',
          })
        }
      })

      await this.fetchPolicies()
      return { succeeded, failed }
    },
  },
})
