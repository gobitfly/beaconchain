<script lang="ts" setup>
import { orderBy } from 'lodash-es'
import {
  type AggregationTimeframe,
  AggregationTimeframes,
  type EfficiencyType,
  EfficiencyTypes,
  SUMMARY_CHART_GROUP_NETWORK_AVERAGE,
  SUMMARY_CHART_GROUP_TOTAL,
  type SummaryChartFilter,
} from '~/types/dashboard/summary'
import { getGroupLabel } from '~/utils/dashboard/group'

const { t: $t } = useTranslation()

const { overview } = useValidatorDashboardOverviewStore()

const chartFilter = defineModel<SummaryChartFilter>({ required: true })

/** aggregation */
const aggregation = ref<AggregationTimeframe>(chartFilter.value.aggregation)

const aggregationList = computed(() => {
  return AggregationTimeframes.map(a => ({
    disabled: (overview.value?.chart_history_seconds?.[a] ?? 0) === 0,
    id: a,
    label: $t(`time_frames.${a}`),
  }))
})

watch(aggregation, (a) => {
  chartFilter.value.aggregation = a
})
const aggregationDisabled = ({ disabled }: { disabled: boolean }) => disabled

/** efficiency */
const efficiency = ref<EfficiencyType>(chartFilter.value.efficiency)

const efficiencyList = EfficiencyTypes.map(e => ({
  id: e,
  label: $t(`dashboard.validator.summary.chart.efficiency.${e}`),
}))
watch(efficiency, (e) => {
  chartFilter.value.efficiency = e
})

/** groups */
const total = ref(
  !chartFilter.value.initialised
  || chartFilter.value.groupIds.includes(SUMMARY_CHART_GROUP_TOTAL),
)
const average = ref(
  !chartFilter.value.initialised
  || chartFilter.value.groupIds.includes(SUMMARY_CHART_GROUP_NETWORK_AVERAGE),
)
const groups = computed(() => {
  if (!overview.value?.groups) {
    return []
  }
  return orderBy(
    overview.value.groups.filter(g => !!g.count),
    [ g => g.name.toLowerCase() ],
    'asc',
  )
})
const selectedGroups = ref<number[]>([])
watch(
  groups,
  (list) => {
    // when groups change we reset the selected groups
    selectedGroups.value = list.map(g => g.id)
  },
  { immediate: true },
)

watch(
  [
    selectedGroups,
    total,
    average,
  ],
  ([
    list,
    t,
    a,
  ]) => {
    const groupIds: number[] = [ ...list ]
    if (t) {
      groupIds.push(SUMMARY_CHART_GROUP_TOTAL)
    }
    if (a) {
      groupIds.push(SUMMARY_CHART_GROUP_NETWORK_AVERAGE)
    }
    chartFilter.value.groupIds = groupIds
    chartFilter.value.initialised = true
  },
  { immediate: true },
)

const selectAllGroups = () => {
  selectedGroups.value = groups.value.map(g => g.id)
}

const toggleGroups = () => {
  if (selectedGroups.value.length < groups.value.length) {
    selectAllGroups()
  }
  else {
    selectedGroups.value = []
  }
}

const selectedLabel = computed(() => {
  const list: string[] = orderBy(
    selectedGroups.value.map(id => getGroupLabel($t, id, groups.value)),
    [ g => g.toLowerCase() ],
    'asc',
  )

  if (average.value) {
    list.splice(0, 0, $t('dashboard.validator.summary.chart.average'))
  }
  if (total.value) {
    list.splice(0, 0, $t('dashboard.validator.summary.chart.total'))
  }
  if (!list.length) {
    return $t('dashboard.group.selection.all')
  }
  if (list.length === 1) {
    return list[0]
  }
  return $t('dashboard.validator.summary.chart.groups', { count: list.length })
})
</script>

<template>
  <div class="chart-filter-row">
    <BcDropdown
      v-model="aggregation" :options="aggregationList" option-value="id" option-label="label"
      :option-disabled="aggregationDisabled" panel-class="summary-chart-aggregation-panel" class="small"
    >
      <template #option="slotProps">
        <span>{{ slotProps.label }}</span>
        <BcPremiumGem class="premium-gem" @click.stop="() => undefined" />
      </template>
    </BcDropdown>
    <BcDropdown v-model="efficiency" :options="efficiencyList" option-value="id" option-label="label" class="small" />

    <BcMultiSelect
      v-model="selectedGroups" class="small" :options="groups" option-label="name" option-value="id"
      :placeholder="$t('dashboard.group.selection.all')"
    >
      <template #header>
        <div class="special-groups">
          <Checkbox v-model="total" input-id="total" :binary="true" />
          <label for="total">{{
            $t("dashboard.validator.summary.chart.total")
          }}</label>
        </div>
        <div class="special-groups">
          <Checkbox v-model="average" input-id="average" :binary="true" />
          <label for="average">{{
            $t("dashboard.validator.summary.chart.average")
          }}</label>
        </div>
        <span class="pointer" @click="toggleGroups">
          {{ $t("dashboard.group.selection.all") }}
        </span>
      </template>
      <template #value>
        {{ selectedLabel }}
      </template>
    </BcMultiSelect>
  </div>
</template>

<style lang="scss" scoped>
.chart-filter-row {
  display: flex;
  gap: var(--padding);

  :deep(> .p-multiselect){
    max-width: 200px;

    @media (max-width: 1000px) {
      max-width: 90px;
    }
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

.special-groups {
  display: flex;
  gap: var(--padding);
  padding-left: var(--padding-small);
  margin-bottom: var(--padding);
}

:global(.summary-chart-aggregation-panel .p-dropdown-item) {
  display: flex;
  gap: var(--padding-small);
  align-items: center;
}

:global(.summary-chart-aggregation-panel .p-dropdown-item:not(.p-disabled) .premium-gem) {
  display: none;
}
</style>
