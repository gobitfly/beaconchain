<script setup lang="ts">
export type BadgeColor = 'gray' | 'green' | 'orange' | 'red'

const {
  color,
} = defineProps<{
  color: BadgeColor,
}>()
</script>

<template>
  <span
    class="bc-badge"
    :class="{
      'bc-badge--gray': color === 'gray',
      'bc-badge--green': color === 'green',
      'bc-badge--orange': color === 'orange',
      'bc-badge--red': color === 'red',
    }"
  >
    <slot />
  </span>
</template>

<style scoped lang="scss">
.bc-badge {
  display: inline-flex;
  justify-content: center;
  border-radius: 4px;
  padding: var(--padding-small) var(--padding);
  text-align: center;
  color: white;
  gap: var(--padding-small);
  font-size: var(--tiny_text_font_size);

  &--gray {
    background-color: var(--grey);
  }

  &--green {
      background-color: var(--positive-color);
  }

  &--orange {
    background-color: var(--orange);
  }

  &--red {
      background-color: var(--negative-color);
  }
}
</style>
