<script setup lang="ts">

import { warn } from 'vue'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { ClElValue } from '~/types/api/common'
import { type OverviewTableData } from '~/types/dashboard/overview'
import { TimeFrames, type NumberOrString } from '~/types/value'
import { totalElClNumbers } from '~/utils/bigMath'

const { t: $t } = useI18n()
const { converter } = useValue()

const tPath = 'dashboard.validator.overview.'

const { validatorDashboardOverview } = storeToRefs(useValidatorDashboardOverviewStore())
warn('DIECE: Validator Overview', validatorDashboardOverview)

const formatValueWei = (value: NumberOrString): NumberOrString => {
  return converter.value.weiToValue(value as string, { fixedDecimalCount: 4 }).label
}
const formatValuePercent = (value: NumberOrString): NumberOrString => {
  return formatPercent(value as number)
}

const createInfo = (key: string, value: ClElValue<number | string>, formatFunction: (value: Partial<NumberOrString>) => NumberOrString) => {
  const clValue = formatFunction(value.cl)
  const elValue = formatFunction(value.el)
  return {
    label: $t(`statistics.${key}`),
    value: `${clValue} (CL) ${elValue} (EL)`
  }
}

const dataList = computed(() => {
  const v = validatorDashboardOverview.value

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

  const onlineClass = v.validators.online ? 'positive' : ''
  const offlineClass = v.validators.online ? 'negative' : ''
  active.value = { label: `<span class="${onlineClass}">${v.validators.online ?? 0}</span> / <span class="${offlineClass}">${v.validators.offline ?? 0}</span>` }
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

  efficiency.value = { label: formatPercent(v.efficiency.last_7d ?? 0) }
  efficiency.infos = TimeFrames.map(k => ({ label: $t(`statistics.${k}`), value: formatValuePercent(v.efficiency[k] ?? 0) }))

  rewards.value = converter.value.weiToValue(totalElCl(v.rewards.last_7d), { addPlus: true })
  rewards.infos = TimeFrames.map(k => createInfo(k, v.rewards[k] ?? 0, formatValueWei))

  apr.value = { label: formatPercent(totalElClNumbers(v.apr.last_7d ?? 0)) }
  apr.infos = TimeFrames.map(k => createInfo(k, v.apr[k] ?? 0, formatValuePercent))
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
