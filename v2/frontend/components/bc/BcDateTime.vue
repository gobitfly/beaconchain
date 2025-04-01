<script setup lang="ts">
import type { Locale } from '~/i18n/i18n.config'

const {
  dateTimeFormat = 'absolute',
  hasDate = true,
  hasTime = true,
} = defineProps<{
  dateTimeFormat?: DateTimeFormat,
  hasDate?: boolean,
  hasTime?: boolean,
  locale?: Locale,
  unixTimestamp: number,
}>()
</script>

<template>
  <span>
    <template v-if="dateTimeFormat === 'absolute'">
      {{ getDateTime(unixTimestamp, {
        hasDate,
        hasTime,
      }) }}
    </template>
    <template v-if="dateTimeFormat === 'relative'">
      {{ getRelativeTime(unixTimestamp) }}
    </template>
  </span>
</template>
