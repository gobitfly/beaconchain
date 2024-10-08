export function logError(message: string) {
  // eslint-disable-next-line no-console
  if (isDevEnvironment) console.error(message)
}
