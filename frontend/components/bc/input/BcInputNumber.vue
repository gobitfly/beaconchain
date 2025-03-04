<script lang="ts" setup generic="WORKAROUND_FOR_CONDITIONAL_PROPS">
// https://github.com/vuejs/core/issues/8952

const {
  inputMode = 'decimal',
  labelPosition = 'left',
} = defineProps<(
  {
    ariaLabel: string,
    label?: never,
  } |
  {
    ariaLabel?: never,
    label: string,
  }
) & {
  inputMode?: 'decimal' | 'numeric',
  inputWidth?: `${number}px` | `${number}rem`,
  labelPosition?: 'left' | 'right',
  max?: number,
  min?: number,
}>()

const inputId = useId()
const inputValue = defineModel<number>({ required: true })
</script>

<template>
  <label
    v-if="label && labelPosition === 'left'"
    :for="inputId"
  >
    {{ label }}
  </label>
  <input
    :id="inputId"
    v-model="inputValue"
    :aria-label
    type="number"
    :inputmode="inputMode"
    :min
    :max
    class="bc-input-number__input"
    v-bind="$attrs"
  >
  <label
    v-if="label && labelPosition === 'right'"
    :for="inputId"
  >
    {{ label }}
  </label>
</template>

<style lang="scss">
.bc-input-number__input {
  --outline-width: 0.125rem;
  --outline-offset: 0.125rem;
  width: v-bind(inputWidth);
  margin: calc(var(--outline-width) + var(--outline-offset));
  border-radius: var(--border-radius);
  background-color: var(--container-background);
  color: var(--container-color);
  border: var(--container-border);
  text-align: center;
  appearance: textfield;

  &::-webkit-outer-spin-button,
  &::-webkit-inner-spin-button {
    -webkit-appearance: none;
  }

  &:focus-visible {
    outline: var(--outline-width) solid var(--blue-500);
    outline-offset: var(--outline-offset);
  }

}
</style>
