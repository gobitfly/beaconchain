<script lang="ts" setup>
const idError = useId()
/**
 * Spacing of error message will be removed by explicitly passing `false`.
 * This should encourage to always think about the error message.
 */
export type BcInputError = false | string
defineProps<{
  error?: BcInputError
}>()
</script>

<template>
  <slot
    :id-error
    :aria-invalid="!!error"
    :aria-describedby="idError"
  />
  <div
    v-if="error !== false"
    :id="idError"
    class="bc-input-error"
  >
    {{ error }}
  </div>
</template>

<style lang="scss" scoped>
// 1. the text is aligned right, but if the message is extra long
// it will start on the left side in the new line
.bc-input-error {
  color: var(--text-color--error);
  grid-column: span 2;
  margin-left: auto; // 1.
  --line-height--default: 1.2;
  min-height: calc(16px * var(--line-height--default));
}
</style>
