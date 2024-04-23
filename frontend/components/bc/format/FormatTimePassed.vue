<script setup lang="ts">
import type { StringUnitLength } from 'luxon'
import { type AgeFormat } from '~/types/settings'
import { formatEpochToDateTime, formatSlotToDateTime } from '~/utils/format'

interface Props {
  value?: number,
  type?: 'epoch' | 'slot', // we can add other types later when needed, we default to epoch
  format?: 'global-setting' | AgeFormat
  noUpdate?: boolean,
  unitLength?: StringUnitLength
}
const props = defineProps<Props>()
const { t: $t } = useI18n()
const { timestamp } = useDate()
const { setting } = useGlobalSetting<AgeFormat>('age-format')

const initTs = ref(timestamp.value) // store the initial timestamp, in case we don't want to auto update

const mappedSetting = computed(() => {
  if (!props.format || props.format === 'global-setting') {
    return setting.value
  }
  return props.format || 'relative'
})

const label = computed(() => {
  if (props.value === undefined) {
    return
  }
  const ts: number = props.noUpdate ? initTs.value : timestamp.value
  switch (props.type) {
    case 'slot':
      return formatSlotToDateTime(props.value, ts, mappedSetting.value, props.unitLength, $t('locales.date'))
    case 'epoch':
    default:
      return formatEpochToDateTime(props.value, ts, mappedSetting.value, props.unitLength, $t('locales.date'))
  }
})
</script>

<template>
  <span v-if="label">{{ label }}</span>
</template>
