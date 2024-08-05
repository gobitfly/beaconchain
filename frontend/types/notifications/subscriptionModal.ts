import type { ChainIDs } from '../network'

export interface InternalEntry {
  type: 'binary' | 'amount' | 'percent' | 'networks'
  networks?: ChainIDs[]
  check?: boolean,
  num?: number
}

export type APIentry = boolean | number | undefined | null | ChainIDs[]
