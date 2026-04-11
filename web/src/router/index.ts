import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import Login from '../views/Login.vue'
import BackendView from '../views/BackendView.vue'
import CaView from '../views/CaView.vue'
import CertificateView from '../views/CertificateView.vue'
import CredentialView from '../views/CredentialView.vue'
import PolicyView from '../views/PolicyView.vue'
import RouteView from '../views/RouteView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    },
    {
      path: '/backends',
      name: 'backends',
      component: BackendView,
    },
    {
      path: '/cas',
      name: 'cas',
      component: CaView,
    },
    {
      path: '/certificates',
      name: 'certificates',
      component: CertificateView,
    },
    {
      path: '/credentials',
      name: 'credentials',
      component: CredentialView,
    },
    {
      path: '/policies',
      name: 'policies',
      component: PolicyView,
    },
    {
      path: '/routes',
      name: 'routes',
      component: RouteView,
    },
  ],
})

// TODO: Enable once auth is implemented
// router.beforeEach(async (to) => {
//   // redirect to login page if not logged in and trying to access a restricted page 
//   const publicPages = ['/login', '/register', '/signup'];
//   const authRequired = !publicPages.includes(to.path);
//   const authStore = useAuthStore();
//   if (authRequired && !authStore.accessToken) {
//     authStore.returnUrl = to.fullPath;
//     return '/login';
//   }
// });

export default router
