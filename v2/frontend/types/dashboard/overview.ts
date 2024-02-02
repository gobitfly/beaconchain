import type { ExtendedLabel } from '~/types/value'

export type DashboardGroup =
  {
    'id': number,
    'name': string
  }

export type DashboardOverview = {
  'groups': DashboardGroup[],
  'validators': {
    'active': number,
    'total': number,
    'pending': number,
    'exited': number,
    'slashed': number
  },
  'efficiency': number
  'rewards': {
    'total': string
    '24h': string
    '7d': string
    '31d': string
    '365d': string
  },
  'luck': {
    'proposal': number
    'sync': number
  },
  'apr': {
    'total': number
    '24h': number
    '7d': number
    '31d': number
    '365d': number
  }
}

export type OverviewTableData = {
  label: string,
  value?: ExtendedLabel,
  additonalValues?: ExtendedLabel[][],
}
