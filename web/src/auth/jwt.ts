type JwtPayload = {
  sub?: string
}

function decodeBase64Url(value: string): string {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padding = normalized.length % 4 === 0 ? '' : '='.repeat(4 - (normalized.length % 4))
  return atob(normalized + padding)
}

export function getJwtSubject(token: string | null | undefined): string | null {
  if (!token) {
    return null
  }

  const parts = token.split('.')
  if (parts.length < 2 || !parts[1]) {
    return null
  }

  try {
    const payload = JSON.parse(decodeBase64Url(parts[1])) as JwtPayload
    return payload.sub ?? null
  } catch {
    return null
  }
}
