<script lang="ts" setup>
import { type ComposerTranslation } from 'vue-i18n'
import { useFormat } from '~/composables/useFormat'
import { useNetworkStore } from '~/stores/useNetworkStore'

interface Props {
  t: ComposerTranslation, // required as dynamically created components via render do not have the proper app context,
  startEpoch: number
}

const props = defineProps<Props>()

const { epochsPerDay } = useNetworkStore()
const { formatEpochToDate } = useFormat()

const dateText = computed(() => {
  const date = formatEpochToDate(props.startEpoch, props.t('locales.date'))
  if (date === undefined) {
    return undefined
  }
  return `${date}`
})

const epochText = computed(() => {
  const endEpoch = props.startEpoch + epochsPerDay()
  return `${props.t('common.epoch')} ${props.startEpoch} - ${endEpoch}`
})
</script>

<template>
  <b>
    <div>
      {{ dateText }}
    </div>
    <div>
      {{ epochText }}
    </div>
  </b>
</template>
