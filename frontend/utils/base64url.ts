export function decodeBase64Url(string: string) {
  string = string
    .replaceAll('-', '+')
    .replaceAll('_', '/')
  return atob(string)
}

export function encodeBase64Url(string: string) {
  return btoa(string)
    .replaceAll('+', '-')
    .replaceAll('/', '_')
    .replaceAll('=', '')
}
