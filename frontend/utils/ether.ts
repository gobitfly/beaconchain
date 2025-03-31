import { REGEXP_PUBLIC_KEX } from './regexp'

export function isPublicKey(value: string): boolean {
  return !!value && REGEXP_PUBLIC_KEX.test(value)
}
