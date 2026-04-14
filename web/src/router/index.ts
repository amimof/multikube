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
      path: '/routes',
      name: 'routes',
      component: () => import('@/views/RoutesView.vue'),
    },
    {
      path: '/cas',
      name: 'cas',
      component: () => import('@/views/CAsView.vue'),
    },
  ],
})

export default router
