<script setup lang="ts">

interface Props {
  variant?: 'default' | 'table'
}

defineProps<Props>()
</script>

<template>
  <Dropdown :class="variant" :panel-class="variant">
    <template #value="slotProps">
      <slot name="value" v-bind="slotProps" />
    </template>
    <template #option="slotProps">
      <slot name="option" v-bind="slotProps.option" />
    </template>
  </Dropdown>
</template>
