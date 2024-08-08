<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import { useNetworkStore } from '~/stores/useNetworkStore'
import {
  type AggregationTimeframe,
  type EfficiencyType,
} from '~/types/dashboard/summary'
import { ONE_HOUR, ONE_DAY, ONE_WEEK } from '~/utils/format'

interface Props {
  t: ComposerTranslation // required as dynamically created components via render do not have the proper app context,
  ts?: number
  startEpoch?: number
  aggregation?: AggregationTimeframe
  efficiencyType?: EfficiencyType
}

const props = defineProps<Props>()

const { tsToEpoch, epochToTs } = useNetworkStore()

const startTs = computed(() => {
  if (props.ts) {
    return props.ts
  }
  if (props.startEpoch) {
    return epochToTs(props.startEpoch)
  }
  return undefined
})

const endTs = computed(() => {
  if (!startTs.value) {
    return
  }
  switch (props.aggregation) {
    case 'epoch':
      return
    case 'hourly':
      return startTs.value + ONE_HOUR
    case 'weekly':
      return startTs.value + ONE_WEEK
    case 'daily':
    default:
      return startTs.value + ONE_DAY
  }
})

const dateText = computed(() => {
  if (!startTs.value) {
    return
  }
  const date = formatGoTimestamp(
    startTs.value,
    undefined,
    'absolute',
    'narrow',
    props.t('locales.date'),
    true,
  )
  if (!endTs.value) {
    return date
  }
  const endDate = formatGoTimestamp(
    endTs.value,
    undefined,
    'absolute',
    'narrow',
    props.t('locales.date'),
    true,
  )

  return `${date} - ${endDate}`
})

const epochText = computed(() => {
  if (!startTs.value) {
    return
  }
  const startEpoch = tsToEpoch(startTs.value)
  if (!endTs.value) {
    return startEpoch
  }
  const endEpoch = tsToEpoch(endTs.value)
  return `${startEpoch} - ${endEpoch}`
})

const title = computed(() => {
  if (props.efficiencyType) {
    return props.t(
      `dashboard.validator.summary.chart.efficiency.${props.efficiencyType}`,
    )
  }
  return undefined
})
</script>

<template>
  <b>
    <div>{{ title }} {{ dateText }}</div>
    <div>{{ t("common.epoch") }} {{ epochText }}</div>
  </b>
</template>
