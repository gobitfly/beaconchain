export function getCSRFHeader (headers?: Headers): [string, string] | undefined {
  if (!headers) {
    return
  }
  for (const entry of headers.entries()) {
    if (entry[0].toUpperCase() === 'X-CSRF-TOKEN') {
      return entry
    }
  }
}
