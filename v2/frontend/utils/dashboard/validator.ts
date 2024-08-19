import type { ValidatorHistoryDuties } from '~/types/api/common'
import type { VDBSummaryValidator } from '~/types/api/validator_dashboard'
import type {
  ValidatorSubset,
  ValidatorSubsetCategory,
} from '~/types/validator'

export function totalDutyRewards(duties?: ValidatorHistoryDuties) {
  if (!duties) {
    return
  }

  const values: string[] = [
    duties.attestation_head?.income,
    duties.attestation_source?.income,
    duties.attestation_target?.income,
    duties.slashing?.income,
    duties.sync?.income,
    duties.proposal?.cl_attestation_inclusion_income,
    duties.proposal?.cl_slashing_inclusion_income,
    duties.proposal?.cl_sync_inclusion_income,
    duties.proposal?.el_income,
  ].filter(v => !!v) as string[]
  if (values.length) {
    return convertSum(...values)
  }
}

export function sortValidatorIds(list?: number[]): number[] {
  if (!list) {
    return []
  }
  return [ ...list ].sort((a, b) => a - b)
}

export function sortSummaryValidators(
  list?: VDBSummaryValidator[],
): VDBSummaryValidator[] {
  if (!list) {
    return []
  }
  return [ ...list ].sort((a, b) => a.index - b.index)
}

export function countSubsetDuties(
  list: ValidatorSubset[],
  categories: ValidatorSubsetCategory[],
): number {
  return categories.reduce((sum, cat) => {
    const subset = list.find(sub => sub.category === cat)
    return sum + countSummaryValidatorDuties(subset?.validators || [], cat)
  }, 0)
}
export function countSummaryValidatorDuties(
  validators: VDBSummaryValidator[],
  category: ValidatorSubsetCategory,
): number {
  let countBy: 'duty-count' | 'duty-value' | 'index' = 'index'
  if (category === 'sync_past') {
    countBy = 'duty-value'
  }
  else if (
    [
      'has_slashed',
      'proposal_missed',
      'proposal_proposed',
    ].includes(category)
  ) {
    countBy = 'duty-count'
  }
  if (countBy === 'index') {
    return validators.length
  }
  return validators.reduce((sum, v) => {
    if (countBy === 'duty-count') {
      return sum + (v.duty_objects?.length || 1)
    }
    return sum + (v.duty_objects?.[0] || 1)
  }, 0)
}
