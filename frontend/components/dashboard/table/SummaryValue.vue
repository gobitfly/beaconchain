<script setup lang="ts">
import { type SummaryDetails, type VDBGroupSummaryResponse, SummaryDetailsEfficiencyProps, type SummaryDetailsEfficiencyProp } from '~/types/dashboard/summary'
import { convertSum } from '~/utils/bigMath';

interface Props {
  property: string,
  detail: SummaryDetails,
  data: VDBGroupSummaryResponse
}
const props = defineProps<Props>()

const data = computed(() => {
  const col = props.data?.[props.detail]
  if (!col) {
    return null
  }
  if (props.property === 'attestation_total') {
    return {
      efficiency: {
        success: col.attestation_head.count.success + col.attestation_source.count.success + col.attestation_target.count.success,
        failed: col.attestation_head.count.failed + col.attestation_source.count.failed + col.attestation_target.count.failed,
        earned: convertSum(col.attestation_head.earned, col.attestation_source.earned, col.attestation_target.earned),
        penality: convertSum(col.attestation_head.penalty, col.attestation_source.penalty, col.attestation_target.penalty)
      }
    }
  } else if (SummaryDetailsEfficiencyProps.includes(props.property as SummaryDetailsEfficiencyProp)) {
    return {
      efficiency: col[props.property as SummaryDetailsEfficiencyProp].count,
      earned: convertSum(col[props.property as SummaryDetailsEfficiencyProp].earned),
      penality: convertSum(col[props.property as SummaryDetailsEfficiencyProp].penalty)
    }
  } else if 
})

</script>
<template>
  <BcTooltip v-if="data?.efficiency">
    <DashboardTableEfficiency :success="data.efficiency.success" :failed="data.efficiency.failed" />
  </BcTooltip>
</template>
