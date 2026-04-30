import { AuthServiceApi, Configuration as AuthConfiguration } from '@/generated/auth'
import { BackendServiceApi, Configuration as BackendConfiguration } from '@/generated/backend'
import { CertificateAuthorityServiceApi, Configuration as CaConfiguration } from '@/generated/ca'
import { CertificateServiceApi, Configuration as CertificateConfiguration } from '@/generated/certificate'
import { CredentialServiceApi, Configuration as CredentialConfiguration } from '@/generated/credential'
import { MetricsServiceApi, Configuration as MetricsConfiguration } from '@/generated/metrics'
import { PolicyServiceApi, Configuration as PolicyConfiguration } from '@/generated/policy'
import { RouteServiceApi, Configuration as RouteConfiguration } from '@/generated/route'
import { UserServiceApi, Configuration as UserConfiguration } from '@/generated/user'
import { clearAuthTokens, getAccessToken, getRefreshToken, setAccessToken } from '@/auth/session'

type ApiMiddleware = {
  pre?: (context: { url: string; init: RequestInit; fetch: typeof fetch }) => Promise<{ url: string; init: RequestInit } | void>
  post?: (context: { url: string; init: RequestInit; response: Response; fetch: typeof fetch }) => Promise<Response | void>
}

const RETRY_HEADER = 'x-multikube-auth-retried'
const AUTH_PATHS = new Set(['/api/v1/auth/login', '/api/v1/auth/refresh'])

let refreshPromise: Promise<string> | null = null

function redirectToLogin() {
  if (typeof window === 'undefined') {
    return
  }

  const loginUrl = new URL('login', window.location.origin + import.meta.env.BASE_URL)
  const currentPath = `${window.location.pathname}${window.location.search}${window.location.hash}`
  if (!currentPath.endsWith('/login') && !currentPath.includes('/login?')) {
    loginUrl.searchParams.set('redirect', `${window.location.pathname}${window.location.search}`)
  }
  window.location.assign(loginUrl.toString())
}

function getPathname(url: string): string {
  if (typeof window === 'undefined') {
    return new URL(url, 'http://localhost').pathname
  }

  return new URL(url, window.location.origin).pathname
}

function isAuthRequest(url: string): boolean {
  return AUTH_PATHS.has(getPathname(url))
}

async function refreshAccessToken(): Promise<string> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    clearAuthTokens()
    redirectToLogin()
    throw new Error('No refresh token available')
  }

  if (!refreshPromise) {
    refreshPromise = authService.authServiceRefresh({
      body: {
        refreshToken,
      },
    })
      .then((response) => {
        if (!response.accessToken) {
          throw new Error('Refresh did not return an access token')
        }

        setAccessToken(response.accessToken)
        return response.accessToken
      })
      .catch((error) => {
        clearAuthTokens()
        redirectToLogin()
        throw error
      })
      .finally(() => {
        refreshPromise = null
      })
  }

  return refreshPromise
}

const authMiddleware: ApiMiddleware = {
  async pre(context) {
    if (isAuthRequest(context.url)) {
      return undefined
    }

    const token = getAccessToken()
    if (!token) {
      return undefined
    }

    const headers = new Headers(context.init.headers ?? {})
    headers.set('Authorization', `Bearer ${token}`)

    return {
      url: context.url,
      init: {
        ...context.init,
        headers,
      },
    }
  },

  async post(context) {
    if (context.response.status !== 401 || isAuthRequest(context.url)) {
      return undefined
    }

    const previousHeaders = new Headers(context.init.headers ?? {})
    if (previousHeaders.get(RETRY_HEADER) === '1') {
      return undefined
    }

    const nextToken = await refreshAccessToken()
    const retryHeaders = new Headers(previousHeaders)
    retryHeaders.set('Authorization', `Bearer ${nextToken}`)
    retryHeaders.set(RETRY_HEADER, '1')

    return context.fetch(context.url, {
      ...context.init,
      headers: retryHeaders,
    })
  },
}

const clientOptions = {
  basePath: '',
  credentials: 'include' as const,
  middleware: [authMiddleware],
}

const authService = new AuthServiceApi(new AuthConfiguration(clientOptions))

export const api = {
  authService,
  backendService: new BackendServiceApi(new BackendConfiguration(clientOptions)),
  certificateAuthorityService: new CertificateAuthorityServiceApi(new CaConfiguration(clientOptions)),
  certificateService: new CertificateServiceApi(new CertificateConfiguration(clientOptions)),
  credentialService: new CredentialServiceApi(new CredentialConfiguration(clientOptions)),
  metricsService: new MetricsServiceApi(new MetricsConfiguration(clientOptions)),
  policyService: new PolicyServiceApi(new PolicyConfiguration(clientOptions)),
  routeService: new RouteServiceApi(new RouteConfiguration(clientOptions)),
  userService: new UserServiceApi(new UserConfiguration(clientOptions)),
}

/**
 * Strip server-managed fields from a resource payload before sending to the API.
 * The generated `V1MetaToJSONTyped` truncates `created`/`updated` to date-only
 * strings which the backend rejects. Rather than editing generated code, we
 * remove those fields (and other server-managed ones) before create/update calls.
 */
export function sanitizePayload<T extends { version?: string; meta?: { name?: string; labels?: Record<string, string> }; config?: unknown }>(
  resource: T,
): { version?: string; meta: { name?: string; labels?: Record<string, string> }; config?: unknown } {
  return {
    version: resource.version,
    meta: {
      name: resource.meta?.name,
      labels: resource.meta?.labels,
    },
    config: resource.config,
  }
}
