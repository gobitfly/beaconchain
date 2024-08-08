import type {
  VDBManageValidatorsTableRow,
  VDBSummaryValidator,
  VDBSummaryValidatorsData,
} from '../api/validator_dashboard'

export type ValidatorStatus = VDBManageValidatorsTableRow['status']

export type ValidatorSubsetCategory =
  | 'all'
  | 'exited_withdrawing'
  | 'exited_withdrawn'
  | 'slashed_withdrawing'
  | 'slashed_withdrawn'
  | VDBSummaryValidatorsData['category']

export type ValidatorSubset = {
  category: ValidatorSubsetCategory
  validators: VDBSummaryValidator[]
}
export type ValidatorSummaryIconRowKey = 'exited' | 'offline' | 'online'

export type SummaryValidatorsIconRowInfo = {
  count: number
  key: ValidatorSummaryIconRowKey
}
