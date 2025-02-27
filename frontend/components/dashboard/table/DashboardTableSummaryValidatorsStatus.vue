<script lang="ts" setup>
defineProps<{
  base?: number | string,
  color: 'gray' | 'green' | 'red',
  value: number | string,
}>()
</script>

<template>
  <span
    class="validator-status"
  >
    <BcIconValidatorStatus
      :color
      width="0.875rem"
    />
    <BaseFormatPercent
      v-if="base"
      :color
      :value
      :base
    />
    <BaseFormatNumber
      v-else
      :color
      :value
    />
  </span>
</template>v

<style scoped lang="scss">
.validator-status {
  display: inline-flex;
  gap: var(--padding-small);
}
</style>
