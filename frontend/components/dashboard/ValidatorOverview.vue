<script setup lang="ts">

import { useValidatorDashboardOverview } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { type OverviewTableData } from '~/types/dashboard/overview'

const { t: $t } = useI18n()

const tPath = 'dashboard.validator.overview.'

// TODO: implement dashboard switching
const { getOverview } = useValidatorDashboardOverview()
await useAsyncData('validator_dashboard_overview', () => getOverview())

const { overview } = storeToRefs(useValidatorDashboardOverview())

const dataList = computed(() => {
  const v = overview.value
  const active: OverviewTableData = {
    label: $t(`${tPath}your_active_validators`)
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
  active.value = `${v.validators.active}/${v.validators.total}`
  active.additonalValues = [
    [
      v.validators.pending ?? 0,
      v.validators.exited ?? 0,
      v.validators.slashed ?? 0
    ],
    [
      $t('validator_state.pending'),
      $t('validator_state.exited'),
      $t('validator_state.slashed')
    ]
  ]
  efficiency.value = formatPercent(v.efficiency)

  rewards.value = formatWeiToEth(v.rewards.total)
  const statsLabels = [
    $t('statistics.24h'),
    $t('statistics.7d'),
    $t('statistics.31d'),
    $t('statistics.365d')
  ]
  rewards.additonalValues = [
    [
      formatWeiToEth(v.rewards['24h']),
      formatWeiToEth(v.rewards['7d']),
      formatWeiToEth(v.rewards['31d']),
      formatWeiToEth(v.rewards['365d'])
    ], statsLabels
  ]

  luck.value = formatPercent(v.luck.proposal)
  luck.additonalValues = [
    [
      formatPercent(v.luck.sync)
    ],
    [
      $t(`${tPath}sync_committee_luck`)
    ]
  ]
  apr.value = formatPercent(v.apr.total)
  apr.additonalValues = [
    [
      formatPercent(v.apr['24h']),
      formatPercent(v.apr['7d']),
      formatPercent(v.apr['31d']),
      formatPercent(v.apr['365d'])
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
.container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  overflow-x: auto;
  gap: 50px;
  height: 101px;
  border: 1px solid var(--container-border-color);
  border-radius: var(--border-radius);
  color: var(--container-color);
  background: var(--container-background);
  padding-left: var(--padding-xl);
  padding-right: var(--padding-xl);
}

.content {
  width: var(--content-width);
  margin: var(--padding) var(--content-margin) var(--padding) var(--content-margin);
}
</style>
