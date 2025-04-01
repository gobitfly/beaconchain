<script setup lang="ts">
const { dateTimeFormat } = storeToRefs(useSettingsStore())

defineProps<{
  unixTimestamp: number,
}>()
</script>

<template>
  <span class="bc-table-date-time">
    <BcDateTime
      :unix-timestamp
      :date-time-format
      :has-time="false"
    />
    <BcDateTime
      v-if="dateTimeFormat === 'absolute'"
      class="bc-table-date-time__time"
      :unix-timestamp
      date-time-format="absolute"
      :has-date="false"
    />
  </span>
</template>

<style scoped lang="scss">
.bc-table-date-time {
  display: flex;
  flex-direction: column;

  .bc-table-date-time__time {
    font-size: 80%;
    color: var(--text-color-discreet);
  }
}
</style>
