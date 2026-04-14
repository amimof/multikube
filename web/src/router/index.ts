import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/backends',
      name: 'backends',
      component: () => import('@/views/BackendsView.vue'),
    },
    {
      path: '/backends/:name',
      name: 'backend-status',
      component: () => import('@/views/BackendStatusView.vue'),
    },
    {
      path: '/routes',
      name: 'routes',
      component: () => import('@/views/RoutesView.vue'),
    },
    {
      path: '/cas',
      name: 'cas',
      component: () => import('@/views/CAsView.vue'),
    },
    {
      path: '/credentials',
      name: 'credentials',
      component: () => import('@/views/CredentialsView.vue'),
    },
    {
      path: '/certificates',
      name: 'certificates',
      component: () => import('@/views/CertificatesView.vue'),
    },
    {
      path: '/policies',
      name: 'policies',
      component: () => import('@/views/PoliciesView.vue'),
    },
  ],
})

export default router
