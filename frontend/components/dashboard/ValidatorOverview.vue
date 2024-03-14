<script setup lang="ts">

import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { DashboardKey } from '~/types/dashboard'
import { type OverviewTableData } from '~/types/dashboard/overview'
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

const dataList = computed(() => {
  const v = overview.value
  const active: OverviewTableData = {
    label: $t(`${tPath}your_online_validators`)
  }
  const efficiency: OverviewTableData = {
    label: $t(`${tPath}total_efficiency`)
  }
  const rewards: OverviewTableData = {
    label: $t(`${tPath}total_rewards`)
  }
  const luck: OverviewTableData = {
    label: $t(`${tPath}proposal_luck`)
  }
  const apr: OverviewTableData = {
    label: $t(`${tPath}total_apr`)
  }
  const list: OverviewTableData[] = [active, efficiency, rewards, luck, apr]
  if (!v) {
    return list
  }
  active.value = { label: `${v.validators.active}/${v.validators.total}` }
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
  efficiency.value = { label: formatPercent(v.efficiency.total) }

  rewards.value = converter.value.weiToValue(totalElCl(v.rewards.all_time), { addPlus: true })
  const statsLabels = [
    { label: `(${$t('statistics.last_24h')})` },
    { label: `(${$t('statistics.last_7d')})` },
    { label: `(${$t('statistics.last_31d')})` },
    { label: `(${$t('statistics.last_365d')})` }
  ]
  rewards.additonalValues = [
    [
      converter.value.weiToValue(totalElCl(v.rewards.last_24h), { addPlus: true }),
      converter.value.weiToValue(totalElCl(v.rewards.last_7d), { addPlus: true }),
      converter.value.weiToValue(totalElCl(v.rewards.last_31d), { addPlus: true }),
      converter.value.weiToValue(totalElCl(v.rewards.last_365d), { addPlus: true })
    ], statsLabels
  ]

  luck.value = { label: formatPercent(v.luck.proposal.percent) }
  luck.additonalValues = [
    [
      { label: formatPercent(v.luck.sync.percent) }
    ],
    [
      { label: $t(`${tPath}sync_committee_luck`) }
    ]
  ]
  apr.value = { label: formatPercent(totalElClNumbers(v.apr.all_time)) }
  apr.additonalValues = [
    [
      { label: formatPercent(totalElClNumbers(v.apr.last_24h)) },
      { label: formatPercent(totalElClNumbers(v.apr.last_7d)) },
      { label: formatPercent(totalElClNumbers(v.apr.last_31d)) },
      { label: formatPercent(totalElClNumbers(v.apr.last_365d)) }
    ], statsLabels
  ]
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
