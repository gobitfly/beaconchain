<script lang="ts" setup>
withDefaults(defineProps<{
  heading: string,
  headingTag: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6',
}>(), {
  heading: '',
})
</script>

<template>
  <div class="rounded border border-gray-300 bg-gray-100 p-5 dark:border-gray-600 dark:bg-gray-800">
    <Component
      :is="headingTag"
      class="mb-5"
    >
      <BaseText variant="2xl">
        {{ heading }}
      </BaseText>
    </Component>
    <slot />
    <footer v-if="$slots.footer" class="mt-3 flex justify-end gap-2">
      <slot name="footer" />
    </footer>
  </div>
</template>

<style lang="scss" scoped>
</style>
