import type { ValidatorHistoryDuties } from '~/types/api/common'
import type { VDBSummaryValidator } from '~/types/api/validator_dashboard'

export function totalDutyRewards (duties?: ValidatorHistoryDuties) {
  if (!duties) {
    return
  }

  const values: string[] = [duties.attestation_head?.income, duties.attestation_source?.income, duties.attestation_target?.income, duties.slashing?.income, duties.sync?.income, duties.proposal?.cl_attestation_inclusion_income, duties.proposal?.cl_slashing_inclusion_income, duties.proposal?.cl_sync_inclusion_income, duties.proposal?.el_income].filter(v => !!v) as string[]
  if (values.length) {
    return convertSum(...values)
  }
}

export function sortValidatorIds (list?: number[]): number[] {
  if (!list) {
    return []
  }
  return [...list].sort((a, b) => a - b)
}

export function sortSummaryValidators (list?: VDBSummaryValidator[]): VDBSummaryValidator[] {
  if (!list) {
    return []
  }
  return [...list].sort((a, b) => a.index - b.index)
}
