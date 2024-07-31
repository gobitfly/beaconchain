<script lang="ts" setup>
import type { VDBGroupSummaryData, VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'
import type { DashboardValidatorContext } from '~/types/dashboard/summary'
import type { ValidatorSubset, ValidatorSubsetCategory } from '~/types/validator'
import { countSubsetDuties } from '~/utils/dashboard/validator'

interface Props {
  context: DashboardValidatorContext,
  subTitle?: string,
  summary?: {
    data?: VDBGroupSummaryData,
    row: VDBSummaryTableRow
  },
  subsets?: ValidatorSubset[]
}
const props = defineProps<Props>()

const infos = computed(() => {
  const list: { value: number | string, slotVizCategory?: SlotVizCategories, className?: string }[] = []
  const percent = {
    total: 0,
    value: 0
  }
  const addSuccessFailed = (category: SlotVizCategories, success?: number, failed?: number, successCategories?: ValidatorSubsetCategory[], failedCategories?: ValidatorSubsetCategory[]) => {
    if (props.subsets?.length && successCategories) {
      success = countSubsetDuties(props.subsets, successCategories)
    }
    if (props.subsets?.length && failedCategories) {
      failed = countSubsetDuties(props.subsets, failedCategories)
    }
    percent.total = (success ?? 0) + (failed ?? 0)
    if (percent.total > 0) {
      if (success !== undefined) {
        percent.value = success
        list.push({ slotVizCategory: category, value: success, className: 'text-positive' })
      }
      if (failed) {
        list.push({ slotVizCategory: category, value: failed, className: 'text-negative' })
      }
    }
  }
  switch (props.context) {
    case 'attestation':
      addSuccessFailed('attestation', props.summary?.row.attestations?.success, props.summary?.row.attestations?.failed)
      break
    case 'sync':
      addSuccessFailed('sync', props.summary?.data?.sync.status_count.success, props.summary?.data?.sync.status_count.failed)
      break
    case 'slashings':
      addSuccessFailed('slashing', props.summary?.data?.slashings.status_count.success, props.summary?.data?.slashings.status_count.failed, ['has_slashed'], ['got_slashed'])
      break
    case 'proposal':
      addSuccessFailed('proposal', props.summary?.row.proposals.success, props.summary?.row.proposals.failed, ['proposal_proposed'], ['proposal_missed'])
      break
    case 'dashboard':
    case 'group':
      if (props.summary?.row.validators) {
        percent.total = props.summary.row.validators.offline + props.summary.row.validators.online
        percent.value = props.summary.row.validators.online
      }
      break
  }

  return { list, percent }
})

</script>

<template>
  <div class="subset-header">
    <span class="sub-title">{{ props?.subTitle }}</span>
    <DashboardTableSummaryValidators
      v-if="props.summary && (context === 'group' || context === 'dashboard')"
      :is-tooltip="true"
      :context="props.context"
      :absolute="true"
      :row="props.summary.row"
    />
    <div v-for="(info, index) in infos.list" :key="index" :class="info.className" class="info">
      <SlotVizIcon v-if="info.slotVizCategory" :icon="info.slotVizCategory" />
      <BcFormatNumber :value="info.value" />
    </div>
    <BcFormatPercent
      v-if="infos.percent.total"
      :base="infos.percent.total"
      :value="infos.percent.value"
      :color-break-point="80"
    />
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";
@use "~/assets/css/utils.scss";

.subset-header {
  display: flex;
  align-items: center;
  gap: var(--padding);
  overflow: hidden;

  .sub-title {
    @include fonts.subtitle_text;
    @include utils.truncate-text
  }

  .info {
    display: flex;
    align-items: center;
    gap: 4px;

    svg {
      height: 14px;
      width: auto;
    }
  }
}
</style>
