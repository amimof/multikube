import { defineStore } from 'pinia'
import { api, sanitizePayload } from '@/api/client'
import type { V1Route } from '@/generated/route'

export const useRouteStore = defineStore('route', {
  state: () => ({
    items: [] as V1Route[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchRoutes() {
      this.loading = true
      this.error = null

      try {
        const response = await api.routeService.routeServiceList()
        this.items = response.routes ?? []
      } catch (err) {
        this.error = err instanceof Error ? err.message : 'Failed to load routes'
        throw err
      } finally {
        this.loading = false
      }
    },

    async createRoute(route: V1Route) {
      const response = await api.routeService.routeServiceCreate({
        route: sanitizePayload(route) as V1Route,
      })
      await this.fetchRoutes()
      return response.route
    },

    async updateRoute(route: V1Route) {
      const name = route.meta?.name

      if (!name) {
        throw new Error('Route is missing name')
      }

      const response = await api.routeService.routeServiceUpdate2({
        name,
        route: sanitizePayload(route) as V1Route,
      })

      await this.fetchRoutes()
      return response.route
    },

    async deleteRoute(route: V1Route) {
      const name = route.meta?.name

      if (!name) {
        throw new Error('Route is missing name')
      }

      await api.routeService.routeServiceDelete2({
        name,
      })

      await this.fetchRoutes()
    },

    async deleteManyRoutes(routes: V1Route[]): Promise<{ succeeded: number; failed: { name: string; error: string }[] }> {
      const results = await Promise.allSettled(
        routes.map(async (r) => {
          const name = r.meta?.name
          if (!name) throw new Error('Route is missing name')
          await api.routeService.routeServiceDelete2({ name })
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
            name: routes[i]?.meta?.name ?? 'unknown',
            error: r.reason instanceof Error ? r.reason.message : 'Delete failed',
          })
        }
      })

      await this.fetchRoutes()
      return { succeeded, failed }
    },
  },
})
