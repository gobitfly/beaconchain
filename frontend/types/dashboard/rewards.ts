import type { BigNumber } from '@ethersproject/bignumber'
import type { ExtendedLabel } from '~/types/value'
import type { ChartSeries } from '~/types/api/common'

export interface RewardChartGroupData {
  id: number;
  name: string;
  bigData: BigNumber[];
}
export interface RewardChartSeries extends ChartSeries<number, number> {
  name: string;
  color: string;
  type: 'bar';
  stack: 'x';
  barMaxWidth: number;
  bigData: BigNumber[];
  formatedData: ExtendedLabel[]
  groups: RewardChartGroupData[];
}
