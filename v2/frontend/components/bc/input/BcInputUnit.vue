<script setup lang="ts">
defineProps<{
  unit: string,
}>()
const input = defineModel<string>()
</script>

<template>
  <span class="bc-input-unit">
    <BcInputText
      v-model="input"
      class="bc-input-unit__input"
      label-position="right"
      input-width="44px"
      :label="unit"
      type="number"
      v-bind="$attrs"
    />
  </span>
</template>

<style scoped lang="scss">
.bc-input-unit {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}
:deep(.bc-input-unit__input) {
  -moz-appearance:textfield;
  appearance: textfield;
  &::-webkit-outer-spin-button,
  &::-webkit-inner-spin-button {
      -webkit-appearance: none;
  }
}
</style>
