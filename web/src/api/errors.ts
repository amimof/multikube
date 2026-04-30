type RpcStatus = {
  message?: string
}

type ResponseLike = {
  response: Response
}

function hasResponse(value: unknown): value is ResponseLike {
  return typeof value === 'object' && value !== null && 'response' in value && value.response instanceof Response
}

export async function getApiErrorMessage(error: unknown, fallback: string): Promise<string> {
  if (hasResponse(error)) {
    try {
      const payload = await error.response.clone().json() as RpcStatus
      if (payload.message && payload.message.length > 0) {
        return payload.message
      }
    } catch {
      // Fall back to the generic error text below.
    }
  }

  if (error instanceof Error && error.message.length > 0 && error.message !== 'Response returned an error code') {
    return error.message
  }

  return fallback
}
