<script lang="ts" setup>
import type {
  VDBGroupSummaryData,
  VDBSummaryTableRow,
} from '~/types/api/validator_dashboard'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'
import type { DashboardValidatorContext } from '~/types/dashboard/summary'
import type {
  SummaryValidatorsIconRowInfo,
  ValidatorSubset,
  ValidatorSubsetCategory,
} from '~/types/validator'
import { countSubsetDuties } from '~/utils/dashboard/validator'

interface Props {
  context: DashboardValidatorContext,
  subsets?: ValidatorSubset[],
  subTitle?: string,
  summary?: {
    data?: VDBGroupSummaryData,
    row: VDBSummaryTableRow,
  },
}
const props = defineProps<Props>()

const infos = computed(() => {
  const validatorIcons: SummaryValidatorsIconRowInfo[] = []
  const list: {
    className?: string,
    slotVizCategory?: SlotVizCategories,
    value: number | string,
  }[] = []
  const percent = {
    total: 0,
    value: 0,
  }
  const addSuccessFailed = (
    category: SlotVizCategories,
    success?: number,
    failed?: number,
    successCategories?: ValidatorSubsetCategory[],
    failedCategories?: ValidatorSubsetCategory[],
  ) => {
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
        list.push({
          className: 'text-positive',
          slotVizCategory: category,
          value: success,
        })
      }
      if (failed) {
        list.push({
          className: 'text-negative',
          slotVizCategory: category,
          value: failed,
        })
      }
    }
  }
  switch (props.context) {
    case 'attestation':
      addSuccessFailed(
        'attestation',
        props.summary?.row.attestations?.success,
        props.summary?.row.attestations?.failed,
      )
      break
    case 'sync':
      addSuccessFailed(
        'sync',
        props.summary?.data?.sync.status_count.success,
        props.summary?.data?.sync.status_count.failed,
      )
      break
    case 'slashings':
      addSuccessFailed(
        'slashing',
        props.summary?.data?.slashings.status_count.success,
        props.summary?.data?.slashings.status_count.failed,
        [ 'has_slashed' ],
        [ 'got_slashed' ],
      )
      break
    case 'proposal':
      addSuccessFailed(
        'proposal',
        props.summary?.row.proposals.success,
        props.summary?.row.proposals.failed,
        [ 'proposal_proposed' ],
        [ 'proposal_missed' ],
      )
      break
    case 'dashboard':
    case 'group': {
      let online = 0
      let offline = 0
      let exited = 0
      if (props.subsets?.length) {
        online = countSubsetDuties(props.subsets, [ 'online' ])
        offline = countSubsetDuties(props.subsets, [ 'offline' ])
        exited = countSubsetDuties(props.subsets, [
          'exited',
          'slashed',
        ])
      }
      else if (props.summary?.row.validators) {
        online = props.summary.row.validators.online
        offline = props.summary.row.validators.offline
        exited = props.summary.row.validators.exited
      }
      if (online) {
        validatorIcons.push({
          count: online,
          key: 'online',
        })
      }
      if (offline) {
        validatorIcons.push({
          count: offline,
          key: 'offline',
        })
      }
      if (exited) {
        validatorIcons.push({
          count: exited,
          key: 'exited',
        })
      }
      // for the total percentage we ignore the exited validators
      percent.total = online + offline
      percent.value = online
    }
  }

  return {
    list,
    percent,
    validatorIcons,
  }
})
</script>

<template>
  <div class="subset-header">
    <span class="sub-title">{{ props?.subTitle }}</span>
    <DashboardTableSummaryValidatorsIconRow
      :icons="infos.validatorIcons"
      :absolute="true"
    />
    <div
      v-for="(info, index) in infos.list"
      :key="index"
      :class="info.className"
      class="info"
    >
      <SlotVizIcon
        v-if="info.slotVizCategory"
        :icon="info.slotVizCategory"
      />
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
    @include utils.truncate-text;
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
