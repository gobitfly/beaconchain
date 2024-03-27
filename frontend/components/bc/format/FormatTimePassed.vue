<script setup lang="ts">
import type { StringUnitLength } from 'luxon'
import { formatEpochToRelative } from '~/utils/format'

interface Props {
  value?: number,
  type?: 'epoch', // we can add slot and other types later when needed, we default to epoch
  noUpdate?: boolean,
  unitLength?: StringUnitLength
}
const props = defineProps<Props>()
const { timestamp } = useDate()

const initTs = ref(timestamp.value) // store the initial timestamp, in case we don't want to auto update

const label = computed(() => {
  if (props.value === undefined) {
    return
  }
  const ts: number = props.noUpdate ? initTs.value : timestamp.value
  switch (props.type) {
    default:
      return formatEpochToRelative(props.value, ts, props.unitLength)
  }
})
</script>
<template>
  <span v-if="label">{{ label }}</span>
</template>
