import { defineStore } from 'pinia';
import { fetch } from '@/utils';
import { computed, ref } from 'vue';

const baseUrl = `${import.meta.env.VITE_API_URL}/auth`;

function parseJwt(token: string) {
  var base64Url = token.split('.')[1] ?? '';
  var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  var jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function (c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));

  return JSON.parse(jsonPayload);
};

interface User {
  id: number;
  email: string;
  name: string;
  roles: string[];
  address: string;
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string>('');
  const refreshToken = ref<string>('');
  const returnUrl = ref<string>('');
  const user = ref<User | null>(null);
  const isLoggedIn = computed(() => !!accessToken.value);
  const isAdmin = computed<boolean>(() => {
    if (user.value != null) {
      for (let i = 0; i < user.value?.roles.length - 1; i++) {
        if (user.value.roles[i] == "admins") {
          return true
        }
      }
    }
    return false
  });
  async function login(username: string, password: string) {
    try {
      const res = await fetch(`${baseUrl}/login`, {
        method: 'POST',
        body: JSON.stringify({ 'username': username, 'password': password })
      });
      accessToken.value = res.body.accessToken;
      refreshToken.value = res.body.refreshToken;
      localStorage.setItem('accessToken', accessToken.value);
      localStorage.setItem('refreshToken', refreshToken.value);

      const jwt = parseJwt(accessToken.value);
      const user = await fetch(`${import.meta.env.VITE_API_URL}/users/${jwt.id}`);
      user.value = user.body.user;
      localStorage.setItem('user', JSON.stringify(user.value));
      return Promise.resolve(accessToken.value);
    } catch (error) {
      return Promise.reject(error);
    }
  }
  function logout() {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
    accessToken.value = '';
    refreshToken.value = '';
    user.value = null;
    window.location.assign('/#/login');
  }

  return {
    accessToken,
    refreshToken,
    returnUrl,
    user,
    isLoggedIn,
    isAdmin,
    login,
    logout,
  }
});
