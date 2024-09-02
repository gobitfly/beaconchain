<script setup lang="ts">
import type { SelectChangeEvent } from 'primevue/select'

interface Props {
  panelClass?: string,
  variant?: 'default' | 'header' | 'table',
}
defineProps<Props>()

const emit = defineEmits<{ (e: 'select', value: SelectChangeEvent['value']): void }>()
</script>

<template>
  <Select
    :class="variant"
    :panel-class="[variant, panelClass]"
    @change="emit('select', $event.value)"
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
      >
        <span
          v-if="slotProps.option.command"
          class="p-select-option-label"
          data-pc-section="optionlabel"
        >
          {{ slotProps.option.label }}
        </span>
      </slot>
    </template>
  </Select>
</template>
