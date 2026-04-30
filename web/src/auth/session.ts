const ACCESS_TOKEN_KEY = 'multikube-auth-access-token'
const REFRESH_TOKEN_KEY = 'multikube-auth-refresh-token'
const USERNAME_KEY = 'multikube-auth-username'

type AuthTokens = {
  accessToken: string | null
  refreshToken: string | null
  username: string | null
}

let accessToken: string | null = null
let refreshToken: string | null = null
let username: string | null = null
let hydrated = false

function storageAvailable(): boolean {
  return typeof window !== 'undefined' && typeof window.localStorage !== 'undefined'
}

function readStoredToken(key: string): string | null {
  if (!storageAvailable()) return null

  const value = window.localStorage.getItem(key)
  return value && value.length > 0 ? value : null
}

function writeStoredToken(key: string, value: string | null) {
  if (!storageAvailable()) return

  if (value && value.length > 0) {
    window.localStorage.setItem(key, value)
    return
  }

  window.localStorage.removeItem(key)
}

export function hydrateAuthSession(): AuthTokens {
  if (!hydrated) {
    accessToken = readStoredToken(ACCESS_TOKEN_KEY)
    refreshToken = readStoredToken(REFRESH_TOKEN_KEY)
    username = readStoredToken(USERNAME_KEY)
    hydrated = true
  }

  return {
    accessToken,
    refreshToken,
    username,
  }
}

export function getAccessToken(): string | null {
  return hydrateAuthSession().accessToken
}

export function getRefreshToken(): string | null {
  return hydrateAuthSession().refreshToken
}

export function getUsername(): string | null {
  return hydrateAuthSession().username
}

export function setAuthTokens(tokens: AuthTokens) {
  hydrated = true
  accessToken = tokens.accessToken
  refreshToken = tokens.refreshToken
  username = tokens.username
  writeStoredToken(ACCESS_TOKEN_KEY, accessToken)
  writeStoredToken(REFRESH_TOKEN_KEY, refreshToken)
  writeStoredToken(USERNAME_KEY, username)
}

export function setAccessToken(token: string | null) {
  setAuthTokens({
    accessToken: token,
    refreshToken: getRefreshToken(),
    username: getUsername(),
  })
}

export function setUsername(value: string | null) {
  setAuthTokens({
    accessToken: getAccessToken(),
    refreshToken: getRefreshToken(),
    username: value,
  })
}

export function clearAuthTokens() {
  setAuthTokens({
    accessToken: null,
    refreshToken: null,
    username: null,
  })
}
