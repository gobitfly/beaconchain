import { warn } from 'vue'

export function toBase64Url (str: string):string {
  return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function fromBase64Url (str: string): string {
  try {
    let res = (str + '=').slice(0, str.length + (str.length % 4))
    res = res.replace(/-/g, '+').replace(/_/g, '/')
    return atob(res)
  } catch (e) {
    warn('error getting fromBase64Url', str)
    return ''
  }
}
