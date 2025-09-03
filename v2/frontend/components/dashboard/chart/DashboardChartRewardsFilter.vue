<script lang="ts" setup>
import type { ChartHistorySeconds } from '~/types/api/common'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

export type RewardsChartFilter = {
  aggregation: AggregationTimeframe,
  before_ts: number,
  group_ids: number[],
}
type AggregationTimeframe = keyof ChartHistorySeconds

const {
  groups,
} = defineProps<{
  groups: VDBOverviewGroup[],
}>()

const { t: $t } = useTranslation()

const { networkInfo } = useNetworkStore()
const validatorDashboardsOverviewStore = useValidatorDashboardOverviewStore()
const {
  hasAbilityRewardsChartHistory,
} = storeToRefs(validatorDashboardsOverviewStore)

const chartFilter = defineModel<RewardsChartFilter>({ required: true })

const aggregationTimeframes: AggregationTimeframe[] = [
  'hourly',
  'daily',
  'weekly',
]
const aggregationList = computed(() => {
  return aggregationTimeframes.map((timeframe) => {
    return {
      disabled: !hasAbilityRewardsChartHistory.value[timeframe],
      id: timeframe,
      label: $t(`time_frames.${timeframe}`),
    }
  })
})
const sortedValidatorGroups = computed(() => {
  return groups?.toSorted(
    (groupA, groupB) => groupA.name.localeCompare(groupB.name, undefined, { sensitivity: 'base' }),
  ) || []
})

const aggregationDisabled = ({ disabled }: { disabled: boolean }) => disabled

const selectedLabel = computed(() => {
  if (!chartFilter.value.group_ids?.length) {
    return $t('dashboard.group.selection.all')
  }
  if (chartFilter.value.group_ids.length === 1) {
    return chartFilter.value.group_ids[0]
  }
  return $t('dashboard.validator.rewards.chart.groups', { count: chartFilter.value.group_ids.length })
})

const handleSetEndDate = (value: Date) => {
  const selectedDateLastMinute = new Date(value).setHours(23, 59, 59, 999)

  chartFilter.value.before_ts = Math.floor(new Date(selectedDateLastMinute).getTime() / 1000)
}
</script>

<template>
  <div class="chart-filter-row">
    <BcDropdown
      v-model="chartFilter.aggregation"
      :options="aggregationList"
      option-value="id"
      option-label="label"
      :option-disabled="aggregationDisabled"
      panel-class="summary-chart-aggregation-panel"
      class="small"
    >
      <template #option="slotProps">
        <div class="summary-chart-aggregation-panel__option">
          <span>{{ slotProps.label }}</span>
          <BcPremiumGem
            v-if="slotProps.disabled"
            class="premium-gem"
            @click.stop="() => undefined"
          />
        </div>
      </template>
    </BcDropdown>

    <BcDatepicker
      id="end-date"
      :label="$t('dashboard.validator.rewards.chart.end_date')"
      :model-value="new Date(chartFilter.before_ts * 1000)"
      :min-date="new Date(networkInfo.timeStampSlot0 * 1000)"
      :max-date="new Date((Date.now()))"
      @update:model-value="handleSetEndDate"
    />

    <BcMultiSelect
      v-if="groups?.length > 1"
      v-model="chartFilter.group_ids"
      class="small"
      :options="sortedValidatorGroups"
      option-label="name"
      option-value="id"
      :placeholder="$t('dashboard.group.selection.all')"
    >
      <template #value>
        {{ selectedLabel }}
      </template>

      <template #header>
        <span>
          {{ $t("dashboard.group.selection.all") }}
        </span>
      </template>
    </BcMultiSelect>
  </div>
</template>

<style lang="scss" scoped>
.chart-filter-row {
  display: flex;
  gap: var(--padding);
  align-items: end;

  :deep(> .p-multiselect){
      max-width: max-content
  }

  :deep(> .p-dropdown) {
    max-width: 200px;

    @media (max-width: 1000px) {
      max-width: 82px;
    }
  }

  @media (max-width: 1000px) {
    gap: var(--padding-small);
  }

}

.premium-gem {
  margin-left: var(--padding-small);
}

:global(.summary-chart-aggregation-panel .p-dropdown-item) {
  display: flex;
  gap: var(--padding-small);
  align-items: center;

  li {
    gap: 1rem;
  }
}
</style>
