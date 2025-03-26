<script setup lang="ts">
defineProps<{
  message?: string,
  title: string,
}>()
</script>

<template>
  <section class="bc-notification-banner">
    <BcText
      is="h2"
      variant="lg"
    >
      {{ title }}
    </BcText>
    <p class="bc-notification-banner__text">
      <slot>
        {{ message }}
      </slot>
    </p>
  </section>
</template>

<style scoped lang="scss">
.bc-notification-banner {
  background-color: var(--primary-orange);
  width: 100%;
  padding: var(--padding) var(--padding-medium);
  color: var(--text-color-inverted);

  &__text {
    margin: 0;
    margin-top: var(--padding-small)
  }

  :deep(.link) {
    color: var(--text-color-inverted);
    text-decoration: underline;
    text-underline-offset: var(--padding-tiny)
  }
}
</style>
