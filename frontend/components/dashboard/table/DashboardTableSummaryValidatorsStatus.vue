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
    <BcColor :color>
      <BcIcon
        name="circle-power-off"
        size="sm"
      />
    </BcColor>
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
