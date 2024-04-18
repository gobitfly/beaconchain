import { warn } from 'vue'

export function toBase64Url (str: string):string {
  if (!str) {
    return ''
  }
  return btoa(str).replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_')
}

export function fromBase64Url (str: string): string {
  if (!str) {
    return ''
  }
  try {
    // eslint-disable-next-line no-useless-escape
    return atob(padBase64(str).replace(/\-/g, '+').replace(/_/g, '/'))
  } catch (e) {
    warn('error getting fromBase64Url', str, e)
    return ''
  }
}

function padBase64 (input:string) {
  const segmentLength = 4
  const stringLength = input.length
  const diff = stringLength % segmentLength

  if (!diff) {
    return input
  }

  let padLength = segmentLength - diff
  let buffer = input

  while (padLength--) {
    buffer += '='
  }

  return buffer.toString()
}
