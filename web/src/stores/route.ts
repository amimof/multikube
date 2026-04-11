import { defineStore } from 'pinia'
import { api } from '@/api/client'
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
			const response = await api.routeService.routeServiceCreate({ route })
			await this.fetchRoutes()
			return response.route
		},

		async updateRoute(route: V1Route) {
			const uid = route.meta?.uid

			if (!uid) {
				throw new Error('Route is missing uid')
			}

			const response = await api.routeService.routeServiceUpdate({
				uid,
				name: route.meta?.name,
				route,
			})

			await this.fetchRoutes()
			return response.route
		},

		async deleteRoute(route: V1Route) {
			const uid = route.meta?.uid

			if (!uid) {
				throw new Error('Route is missing uid')
			}

			await api.routeService.routeServiceDelete({
				uid,
				name: route.meta?.name,
			})

			await this.fetchRoutes()
		},
	},
})
