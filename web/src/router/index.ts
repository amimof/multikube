import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    publicOnly?: boolean
  }
}

const protectedChildren = [
  {
    path: '',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
  },
  {
    path: 'backends',
    name: 'backends',
    component: () => import('@/views/BackendsView.vue'),
  },
  {
    path: 'backends/:name',
    name: 'backend-status',
    component: () => import('@/views/BackendStatusView.vue'),
  },
  {
    path: 'routes',
    name: 'routes',
    component: () => import('@/views/RoutesView.vue'),
  },
  {
    path: 'routes/:name',
    name: 'route-status',
    component: () => import('@/views/RouteStatusView.vue'),
  },
  {
    path: 'cas',
    name: 'cas',
    component: () => import('@/views/CAsView.vue'),
  },
  {
    path: 'cas/:name',
    name: 'ca-status',
    component: () => import('@/views/CAStatusView.vue'),
  },
  {
    path: 'credentials',
    name: 'credentials',
    component: () => import('@/views/CredentialsView.vue'),
  },
  {
    path: 'credentials/:name',
    name: 'credential-status',
    component: () => import('@/views/CredentialStatusView.vue'),
  },
  {
    path: 'certificates',
    name: 'certificates',
    component: () => import('@/views/CertificatesView.vue'),
  },
  {
    path: 'certificates/:name',
    name: 'certificate-status',
    component: () => import('@/views/CertificateStatusView.vue'),
  },
  {
    path: 'policies',
    name: 'policies',
    component: () => import('@/views/PoliciesView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { publicOnly: true },
    },
    {
      path: '/',
      component: () => import('@/layouts/AppShell.vue'),
      meta: { requiresAuth: true },
      children: protectedChildren,
    },
  ],
})

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (!authStore.initialized) {
    authStore.restoreSession()
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  if (to.meta.publicOnly && authStore.isAuthenticated) {
    const redirect = typeof to.query.redirect === 'string' && to.query.redirect.length > 0
      ? to.query.redirect
      : '/'

    return redirect
  }

  return true
})

export default router
