<script setup lang="ts">
withDefaults(
  defineProps<{
    tag?: 'h2' | 'p' | 'span',
    variant?: 'base' | 'lg',
  }>(), {
    tag: 'span',
    variant: 'base',
  },
)
</script>

<template>
  <component
    :is="tag"
    class="reset"
    :class="{
      'variant-lg': variant === 'lg',
    }"
  >
    <slot />
  </component>
</template>

<style scoped lang="scss">
@use '~/assets/css/breakpoints' as *;
.reset {
  margin-block: unset;
}
.variant-lg {
  font-family: "Montserrat";
  font-size: 1.25rem;
  font-weight: 500;
  line-height: 1.2;

  @media (min-width: $breakpoint-sm) {
    font-size: 1.5rem;
  }
}
</style>
