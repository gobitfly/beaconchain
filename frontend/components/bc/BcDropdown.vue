<script setup lang="ts">
interface Props {
  panelClass?: string,
  variant?: 'default' | 'header' | 'table',
}

defineProps<Props>()
</script>

<template>
  <Select
    :class="variant"
    :panel-class="[variant, panelClass]"
  >
    <template #dropdownicon>
      <IconChevron direction="bottom" />
    </template>
    <template #value="slotProps">
      <slot
        name="value"
        v-bind="slotProps"
      />
    </template>
    <template #option="slotProps">
      <slot
        name="option"
        v-bind="slotProps.option"
      />
    </template>
  </Select>
</template>
