import type { BigNumber } from '@ethersproject/bignumber'
import type { ExtendedLabel } from '~/types/value'
import type { ChartSeries } from '~/types/api/common'

export interface RewardChartGroupData {
  bigData: BigNumber[]
  id: number
  name: string
}
export interface RewardChartSeries extends ChartSeries<number, number> {
  barMaxWidth: number
  bigData: BigNumber[]
  color: string
  formatedData: ExtendedLabel[]
  groups: RewardChartGroupData[]
  name: string
  stack: 'x'
  type: 'bar'
}
