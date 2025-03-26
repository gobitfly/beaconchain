<script setup lang="ts">
withDefaults(
  defineProps<{
    is?:
      | 'h2'
      | 'h4'
      | 'p'
      | 'span',
    isDimmed?: boolean,
    variant?: 'base' | 'lg' | 'md' | 'sm',
  }>(), {
    is: 'span',
    variant: 'base',
  },
)
</script>

<template>
  <component
    :is
    :class="{
      'variant-lg': variant === 'lg',
      'variant-md': variant === 'md',
      'variant-sm': variant === 'sm',
      'is-dimmed': isDimmed,
    }"
  >
    <slot />
  </component>
</template>

<style scoped lang="scss">
@use '~/assets/css/breakpoints' as *;
.is-dimmed {
  opacity: 0.7;
}

.variant-sm {
  font-size: 0.875rem;
}
.variant-md {
  font-family: "Montserrat";
  font-size: 1.25rem;
  font-weight: 500;
  line-height: 1.2;
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
