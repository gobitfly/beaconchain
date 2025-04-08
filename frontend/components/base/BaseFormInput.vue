<script setup lang="ts">
const input = defineModel<string>()
const idInput = useId()
const idError = useId()
defineProps<{
  /**
   * Spacing of error message will be removed by explicitly passing `false`.
   * This should encourage to always think about the error message and question this.
   */
  errorMessage?: false | string,
  /**
   * <label> will be removed by explicitly passing `false`.
   * This should encourage to always think about a proper label and question this.
   * ♿️ When removing the label, you should make sure the input is still accessible.
   */
  label: false | string,
  type: HTMLInputElement['type'],
}>()
</script>

<template>
  <BaseGutter>
    <label v-if="label !== false" :for="idInput">
      {{ label }}
    </label>
    <input
      :id="idInput"
      v-model="input"
      :placeholder="$t('common.placeholder.email')"
      class="rounded border border-gray-300 bg-gray-200 px-2 py-1 text-gray-400 ring-orange-400 focus:outline-none focus-visible:ring-2 dark:border-gray-600 dark:bg-gray-700"
      :aria-invalid="!!errorMessage"
      :aria-describedby="errorMessage ? idError : undefined"
      :type
      v-bind="$attrs"
    >
    <BaseFormError v-if="errorMessage !== false" :id="idError">
      {{ errorMessage }}
    </BaseFormError>
  </BaseGutter>
</template>

<style scoped></style>
