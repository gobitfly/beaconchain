<script lang="ts" setup>
import type { BcInputError } from '~/components/bc/input/BcInputError.vue'

const idInput = useId()
const input = defineModel<string>()

const props = defineProps<{
  error?: BcInputError,
  inputWidth?: `${number}px`,
  label: string,
  placeholder?: string,
  shouldAutoselect?: boolean,
  type?: HTMLInputElement['type'],
}>()
onMounted(() => {
  if (props.shouldAutoselect) {
    const input = document.getElementById(idInput)
    if (input instanceof HTMLInputElement) {
      input.focus()
      input.select()
    }
  }
})
</script>

<template>
  <BcInputError :error>
    <label :for="idInput">{{ label }}</label>
    <InputText
      :id="idInput"
      v-model.trim="input"
      v-bind="$attrs"
      class="bc-input-text__input"
      :placeholder
      :type
    />
  </BcInputError>
</template>

<style lang="scss">
.bc-input-text__input {
  width: v-bind(inputWidth);
}
</style>
