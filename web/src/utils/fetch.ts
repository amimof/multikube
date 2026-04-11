// @ts-nocheck
import merge from 'lodash/merge'
import { useAuthStore } from '@/stores';
import { configureRefreshFetch, fetchJSON } from 'refresh-fetch';

const retrieveToken = () => localStorage.getItem('accessToken');
const saveToken = token => {
  const store = useAuthStore();
  localStorage.setItem('accessToken', token);
  store.accessToken = token;
}
const clearToken = () => localStorage.removeItem('accessToken');
const clearRefreshToken = () => localStorage.removeItem('refreshToken');
const retrieveRefreshToken = () => localStorage.getItem('refreshToken');
const logout = () => {
  const store = useAuthStore();
  store.logout();
}

const fetchJSONWithToken = (url, options = {}) => {
  const token = retrieveToken()

  let optionsWithToken = options
  if (token != null) {
    optionsWithToken = merge({}, options, {
      headers: {
        Authorization: `${token}`
      }
    })
  }

  return fetchJSON(url, optionsWithToken)
}

const shouldRefreshToken = error =>
  error.response.status === 401


const refreshToken = () => {
  const token = retrieveRefreshToken()
  return fetchJSONWithToken(`${import.meta.env.VITE_API_URL}/auth/refresh`, {
    method: 'POST',
    body: JSON.stringify({
      refreshToken: `${token}`
    })
  })
    .then(response => {
      saveToken(response.body.accessToken)
    })
    .catch(error => {
      // Clear token and continue with the Promise catch chain
      logout();
      //throw error
    })
}

const fetch = configureRefreshFetch({
  fetch: fetchJSONWithToken,
  shouldRefreshToken,
  refreshToken
});

export {
  fetch,
}

export const fetchWrapper = {
  get: request('GET'),
  post: request('POST'),
  put: request('PUT'),
  delete: request('DELETE'),
  patch: request('PATCH'),
};

function request(method) {
  return (url, body, headers) => {
    const auth = useAuthStore();
    const requestOptions = {
      method,
      headers: headers
    };
    const isLoggedIn = !!auth?.token;
    const isApiUrl = url.startsWith(import.meta.env.VITE_API_URL);
    const publicPages = ['/api/v1/auth/token', '/api/v1/auth/refresh'];
    const u = new URL(url)
    const authRequired = !publicPages.includes(u.pathname);

    if (isLoggedIn && isApiUrl && authRequired) {
      requestOptions.headers['Authorization'] = `${auth.token}`
    }
    if (body) {
      requestOptions.headers['Content-Type'] = 'application/json';
      requestOptions.body = JSON.stringify(body);
    }
    return fetch_retry(url, requestOptions)
  }
}

function fetch_retry(url, option, n) {
  return new Promise((resolve, reject) => {
    fetch(url, option).then((response) => {
      if (!response.ok) {
        const { user, logout, refresh } = useAuthStore();
        if ([401].includes(response.status) && user) {
          refresh().then(() => {
            if (n === 1) return resolve(response);
            fetch_retry(url, option, n - 1).then(resolve).catch(reject);
          }).catch(err => {
            logout();
            window.location = '/#/login';
            reject(err)
          })
        }
      }
      const d = handleResponse(response);
      return resolve(d);
    }).catch((error) => {
      return reject(error);
    })
  })
}

async function handleResponse(response) {
  const isJson = response.headers?.get('content-type')?.includes('application/json');
  const data = isJson ? await response.json() : null;
  return Promise.resolve(data);
}
