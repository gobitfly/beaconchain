import type { ChainIDs } from '../network'

export interface InternalEntry {
  check?: boolean,
  networks?: ChainIDs[],
  num?: number,
  type: 'amount' | 'binary' | 'networks' | 'percent',
}

export type APIentry = boolean | ChainIDs[] | null | number | undefined
