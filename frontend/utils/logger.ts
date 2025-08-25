export function logError(message: string,
  details: {
    error?: unknown,
    payload?: any,
  } = {}) {
  // eslint-disable-next-line no-console
  console.error(message, details)
}
