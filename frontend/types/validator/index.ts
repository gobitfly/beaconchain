import type { VDBManageValidatorsTableRow, VDBSummaryValidator, VDBSummaryValidatorsData } from '../api/validator_dashboard'

export type ValidatorStatus = VDBManageValidatorsTableRow['status']

export type ValidatorSubsetCategory = VDBSummaryValidatorsData['category'] | 'all' | 'exited_withdrawing' | 'exited_withdrawn' | 'slashed_withdrawing' | 'slashed_withdrawn'

export type ValidatorSubset = {
  category: ValidatorSubsetCategory
  validators: VDBSummaryValidator[]
}
export type ValidatorSummaryIconRowKey = 'online' | 'offline' | 'exited'

export type SummaryValidatorsIconRowInfo = {
  count: number
  key: ValidatorSummaryIconRowKey
}
