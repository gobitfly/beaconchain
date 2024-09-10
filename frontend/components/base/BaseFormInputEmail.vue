<script setup lang="ts">
const props = defineProps<{
  /**
   * fieldName is used to validate the field
   */
  fieldName?: string,
  label: false | string,
}>()

const { t: $t } = useTranslation()

const {
  errorMessage,
  validate,
  value,
} = useField<string>(
  () => props.fieldName ?? '',
  validation.email(
    $t('validation.email.invalid'),
    { isRequiredMessage: $t('validation.email.empty') },
  ),
  { validateOnValueUpdate: false },
)
</script>

<template>
  <BaseFormInput
    v-model="value"
    :label
    type="email"
    :error-message
    v-bind="$attrs"
    @blur="validate"
  />
</template>

<style scoped></style>
