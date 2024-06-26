import type { VDBManageValidatorsTableRow, VDBSummaryValidator, VDBSummaryValidatorsData } from '../api/validator_dashboard'

export type ValidatorStatus = VDBManageValidatorsTableRow['status']

export type ValidatorSubsetCategory = VDBSummaryValidatorsData['category'] | 'all'

export type ValidatorSubset = {
  category: ValidatorSubsetCategory,
  validators: VDBSummaryValidator[]
}
