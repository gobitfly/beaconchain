<script setup lang="ts">
import type { StringUnitLength } from 'luxon'
import { useFormat } from '~/composables/useFormat'
import { type AgeFormat } from '~/types/settings'
import { formatGoTimestamp } from '~/utils/format'

const {
  formatEpochToDateTime, formatSlotToDateTime,
} = useFormat()

interface Props {
  format?: 'global-setting' | AgeFormat,
  noUpdate?: boolean,
  type?: 'epoch' | 'go-timestamp' | 'slot', // we can add other types later when needed, we default to epoch
  unitLength?: StringUnitLength,
  value?: number | string,
}
const props = defineProps<Props>()
const { t: $t } = useTranslation()
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
  let text: null | string | undefined = ''
  switch (props.type) {
    case 'go-timestamp':
      text = formatGoTimestamp(
        props.value,
        ts,
        mappedSetting.value,
        props.unitLength,
        $t('locales.date'),
      )
      break
    case 'slot':
      text = formatSlotToDateTime(
        props.value as number,
        ts,
        mappedSetting.value,
        props.unitLength,
        $t('locales.date'),
      )
      break
    case 'epoch':
    default:
      text = formatEpochToDateTime(
        props.value as number,
        ts,
        mappedSetting.value,
        props.unitLength,
        $t('locales.date'),
      )
  }

  if (text && mappedSetting.value === 'absolute') {
    const lastComma = text.lastIndexOf(',')
    if (lastComma > 0) {
      return {
        subtext: text.slice(lastComma + 1),
        text: text.slice(0, lastComma),
      }
    }
  }

  return { text }
})
</script>

<template>
  <span
    v-if="label"
    class="text"
  >
    <div>{{ label.text }}</div>
    <div
      v-if="label.subtext"
      class="subtext"
    >{{ label.subtext }}</div>
  </span>
</template>

<style lang="scss" scoped>
.text {
  display: flex;
  flex-direction: column;

  .subtext {
    font-size: 80%;
    color: var(--text-color-discreet);
  }
}
</style>
