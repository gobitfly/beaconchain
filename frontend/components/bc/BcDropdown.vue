<script setup lang="ts">

interface Props {
  variant?: 'default' | 'table' | 'header',
  panelClass?: string
}

defineProps<Props>()
</script>

<template>
  <Dropdown :class="variant" :panel-class="[variant, panelClass]">
    <template #value="slotProps">
      <slot name="value" v-bind="slotProps" />
    </template>
    <template #option="slotProps">
      <slot name="option" v-bind="slotProps.option" />
    </template>
  </Dropdown>
</template>
