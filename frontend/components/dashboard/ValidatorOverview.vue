<script setup lang="ts">

import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { ClElValue } from '~/types/api/common'
import type { DashboardKey } from '~/types/dashboard'
import { type OverviewTableData } from '~/types/dashboard/overview'
import type { PeriodicValuesKey } from '~/types/value'
import { totalElClNumbers } from '~/utils/bigMath'

interface Props {
  dashboardKey: DashboardKey
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const { converter } = useValue()

const tPath = 'dashboard.validator.overview.'

// TODO: implement dashboard switching
const { getOverview } = useValidatorDashboardOverviewStore()
await useAsyncData('validator_dashboard_overview', () => getOverview(props.dashboardKey))

watch(() => props.dashboardKey, () => {
  getOverview(props.dashboardKey)
}, { immediate: true })

const { overview } = storeToRefs(useValidatorDashboardOverviewStore())

const formatInfoValue = (value: string | number): string | number => {
  if (typeof value === 'string') {
    return converter.value.weiToValue(value, { fixedDecimalCount: 4 }).label
  }
  return formatPercent(value as number)
}

const createInfo = (key: string, value: ClElValue<number | string>) => {
  const clValue = formatInfoValue(value.cl)
  const elValue = formatInfoValue(value.el)
  return {
    label: $t(`statistics.${key}`),
    value: `${clValue} (CL) ${elValue} (EL)`
  }
}

const dataList = computed(() => {
  const v = overview.value
  const active: OverviewTableData = {
    label: $t(`${tPath}your_online_validators`)
  }
  const efficiency: OverviewTableData = {
    label: $t(`${tPath}7d_efficiency`)
  }
  const rewards: OverviewTableData = {
    label: $t(`${tPath}7d_rewards`)
  }
  const apr: OverviewTableData = {
    label: $t(`${tPath}7d_apr`)
  }
  const list: OverviewTableData[] = [active, efficiency, rewards, apr]
  if (!v) {
    return list
  }
  active.value = { label: `${v.validators.online}/${v.validators.offline}` }
  active.additonalValues = [
    [
      { label: v.validators.pending ?? 0 },
      { label: v.validators.exited ?? 0 },
      { label: v.validators.slashed ?? 0 }
    ],
    [
      { label: $t('validator_state.pending') },
      { label: $t('validator_state.exited') },
      { label: $t('validator_state.slashed') }
    ]
  ]
  const keys: PeriodicValuesKey[] = ['last_24h', 'last_7d', 'last_30d', 'all_time']

  efficiency.value = { label: formatPercent(v.efficiency.last_7d) }
  efficiency.infos = keys.map(k => ({ label: $t(`statistics.${k}`), value: formatInfoValue(v.efficiency[k]) }))

  rewards.value = converter.value.weiToValue(totalElCl(v.rewards.last_7d), { addPlus: true })
  rewards.infos = keys.map(k => createInfo(k, v.rewards[k]))

  apr.value = { label: formatPercent(totalElClNumbers(v.apr.last_7d)) }
  apr.infos = keys.map(k => createInfo(k, v.apr[k]))
  return list
})

</script>
<template>
  <div class="container">
    <DashboardOverviewBox v-for="data in dataList" :key="data.label" :data="data" />
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
.container {
  @include main.container;
  display: flex;
  align-items: center;
  justify-content: space-between;
  overflow-x: auto;
  gap: 50px;
  height: 101px;
  padding-left: var(--padding-xl);
  padding-right: var(--padding-xl);
}

.content {
  width: var(--content-width);
  margin: var(--padding) var(--content-margin) var(--padding) var(--content-margin);
}
</style>
