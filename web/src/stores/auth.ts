import { defineStore } from 'pinia'
import { api } from '@/api/client'
import { getApiErrorMessage } from '@/api/errors'
import { getJwtSubject } from '@/auth/jwt'
import { clearAuthTokens, getAccessToken, getRefreshToken, getUsername, hydrateAuthSession, setAccessToken, setAuthTokens, setUsername } from '@/auth/session'

type LoginCredentials = {
  username: string
  password: string
}

function requireToken(value: string | undefined, message: string): string {
  if (!value) {
    throw new Error(message)
  }

  return value
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: null as string | null,
    refreshToken: null as string | null,
    username: null as string | null,
    initialized: false,
    loginLoading: false,
    error: null as string | null,
    refreshPromise: null as Promise<string> | null,
  }),

  getters: {
    isAuthenticated: (state) => Boolean(state.accessToken && state.refreshToken),
  },

  actions: {
    restoreSession() {
      const session = hydrateAuthSession()
      this.accessToken = session.accessToken
      this.refreshToken = session.refreshToken
      this.username = session.username ?? getJwtSubject(session.accessToken)
      this.initialized = true
    },

    setSession(accessToken: string, refreshToken: string) {
      const username = getJwtSubject(accessToken)
      setAuthTokens({ accessToken, refreshToken, username })
      this.accessToken = accessToken
      this.refreshToken = refreshToken
      this.username = username
      this.error = null
      this.initialized = true
    },

    updateAccessToken(accessToken: string) {
      const username = getJwtSubject(accessToken) ?? this.username
      setAccessToken(accessToken)
      setUsername(username)
      this.accessToken = accessToken
      this.username = username
      this.initialized = true
    },

    clearSession() {
      clearAuthTokens()
      this.accessToken = null
      this.refreshToken = null
      this.username = null
      this.error = null
      this.refreshPromise = null
      this.initialized = true
    },

    async login(credentials: LoginCredentials) {
      this.loginLoading = true
      this.error = null

      try {
        const response = await api.authService.authServiceLogin({
          body: {
            username: credentials.username,
            password: credentials.password,
          },
        })

        this.setSession(
          requireToken(response.accessToken, 'Login did not return an access token'),
          requireToken(response.refreshToken, 'Login did not return a refresh token'),
        )
      } catch (error) {
        this.error = await getApiErrorMessage(error, 'Login failed')
        throw error
      } finally {
        this.loginLoading = false
      }
    },

    async refreshAccessToken() {
      const currentRefreshToken = this.refreshToken ?? getRefreshToken()
      if (!currentRefreshToken) {
        this.clearSession()
        throw new Error('No refresh token available')
      }

      if (!this.refreshPromise) {
        this.refreshPromise = (async () => {
          const response = await api.authService.authServiceRefresh({
            body: {
              refreshToken: currentRefreshToken,
            },
          })

          const nextAccessToken = requireToken(response.accessToken, 'Refresh did not return an access token')
          this.updateAccessToken(nextAccessToken)
          return nextAccessToken
        })()

        this.refreshPromise = this.refreshPromise
          .catch((error) => {
            this.clearSession()
            throw error
          })
          .finally(() => {
            this.refreshPromise = null
          })
      }

      return this.refreshPromise
    },

    logout() {
      this.clearSession()
    },

    syncFromSession() {
      this.accessToken = getAccessToken()
      this.refreshToken = getRefreshToken()
      this.username = getUsername() ?? getJwtSubject(this.accessToken)
      this.initialized = true
    },
  },
})
