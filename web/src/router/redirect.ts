const basePath = import.meta.env.BASE_URL.replace(/\/$/, '')

export function normalizeRedirectPath(value: string | null | undefined): string {
  if (!value || value.length === 0) {
    return '/'
  }

  let normalized = value

  if (/^https?:\/\//.test(normalized)) {
    try {
      const url = new URL(normalized)
      normalized = `${url.pathname}${url.search}${url.hash}`
    } catch {
      return '/'
    }
  }

  if (!normalized.startsWith('/')) {
    normalized = `/${normalized}`
  }

  if (basePath && normalized === basePath) {
    return '/'
  }

  if (basePath && normalized.startsWith(`${basePath}/`)) {
    const suffix = normalized.slice(basePath.length)
    return suffix.length > 0 ? suffix : '/'
  }

  return normalized
}
