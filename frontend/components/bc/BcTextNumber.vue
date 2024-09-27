<script setup lang="ts">
/**
 * Should prevent layout shift, e.g. when the value changes from 1 to 2 or 3 to 4 digits
 *
 * @example `10` should occuy the same space as ` 9`
 */
defineProps<{
  minWidth: `${number}ch`,
  prefix?: string,
  suffix?: string,
}>()
</script>

<template>
  <span class="bc-text-number">
    <span v-if="prefix">
      {{ prefix }}
    </span>
    <span class="bc-text-number__value">
      <slot />
    </span>
    <span v-if="suffix">
      {{ suffix }}
    </span>
  </span>
</template>

<style scoped lang="scss">
.bc-text-number {
  white-space: nowrap;
}
.bc-text-number__value {
  display: inline-block;
  min-width: v-bind(minWidth);
  text-align: end;
}
</style>
