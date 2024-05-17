<script setup lang="ts">
import type { ValidatorHistoryDuties } from '~/types/api/common'
import { formatRewardValueOption } from '~/utils/dashboard/table'
import { totalDutyRewards } from '~/utils/dashboard/validator'

interface Props {
  data?: ValidatorHistoryDuties,
}
const props = defineProps<Props>()

const { t: $t } = useI18n()

const mapped = computed(() => {
  const total = totalDutyRewards(props.data)
  const details: {label: string, value?: string}[] = []
  if (!total || total.isZero()) {
    return {
      total,
      details
    }
  }

  const addDetail = (key: string, value?: string) => {
    if (!value || value === '0') {
      return
    }
    details.push({
      label: $t(`validator.rewards.${key}`),
      value
    })
  }

  addDetail('attestation_head', props?.data?.attestation_head?.income)
  addDetail('attestation_source', props?.data?.attestation_source?.income)
  addDetail('attestation_target', props?.data?.attestation_target?.income)
  addDetail('proposer_el', props?.data?.proposal?.el_income)
  addDetail('proposer_attestation', props?.data?.proposal?.cl_attestation_inclusion_income)
  addDetail('proposer_sync', props?.data?.proposal?.cl_sync_inclusion_income)
  addDetail('proposer_slashing', props?.data?.proposal?.cl_slashing_inclusion_income)
  addDetail('total', total.toString())
  return {
    total,
    details
  }
})

</script>
<template>
  <BcFormatValue
    :value="mapped.total"
    :use-colors="true"
    :options="formatRewardValueOption"
  >
    <template v-if="mapped.details?.length" #tooltip>
      <div class="tooltip">
        <div v-for="detail in mapped.details" :key="detail.label">
          <b>{{ detail.label }}: </b>
          <BcFormatValue
            :value="detail.value"
            :use-colors="true"
            :options="formatRewardValueOption"
          />
        </div>
      </div>
    </template>
  </BcFormatValue>
</template>
<style lang="scss" scoped>
.tooltip{
  text-align: left;
  .head{
    margin-top: var(--padding);
  }
}
</style>
