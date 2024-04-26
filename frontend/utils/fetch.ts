export function getCSRFHeader (headers?: Headers): [string, string] | undefined {
  if (!headers) {
    return
  }
  for (const entry of headers.entries()) {
    if (entry[0].includes('X-CSRF-Token')) {
      return entry
    }
  }
}
