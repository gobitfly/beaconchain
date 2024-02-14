<script setup lang="ts">

import { useValidatorDashboardOverview } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { type OverviewTableData } from '~/types/dashboard/overview'

const { t: $t } = useI18n()
const { converter } = useValue()

const tPath = 'dashboard.validator.overview.'

// TODO: implement dashboard switching
const { getOverview } = useValidatorDashboardOverview()
await useAsyncData('validator_dashboard_overview', () => getOverview())

const { overview } = storeToRefs(useValidatorDashboardOverview())

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
  efficiency.value = { label: formatPercent(v.efficiency) }

  rewards.value = converter.value.weiToValue(v.rewards.total, { addPlus: true })
  const statsLabels = [
    { label: $t('statistics.24h') },
    { label: $t('statistics.7d') },
    { label: $t('statistics.31d') },
    { label: $t('statistics.365d') }
  ]
  rewards.additonalValues = [
    [
      converter.value.weiToValue(v.rewards['24h'], { addPlus: true }),
      converter.value.weiToValue(v.rewards['7d'], { addPlus: true }),
      converter.value.weiToValue(v.rewards['31d'], { addPlus: true }),
      converter.value.weiToValue(v.rewards['365d'], { addPlus: true })
    ], statsLabels
  ]

  luck.value = { label: formatPercent(v.luck.proposal) }
  luck.additonalValues = [
    [
      { label: formatPercent(v.luck.sync) }
    ],
    [
      { label: $t(`${tPath}sync_committee_luck`) }
    ]
  ]
  apr.value = { label: formatPercent(v.apr.total) }
  apr.additonalValues = [
    [
      { label: formatPercent(v.apr['24h']) },
      { label: formatPercent(v.apr['7d']) },
      { label: formatPercent(v.apr['31d']) },
      { label: formatPercent(v.apr['365d']) }
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
