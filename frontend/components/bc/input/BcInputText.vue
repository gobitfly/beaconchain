<script lang="ts" setup>
const idInput = useId()
const input = defineModel<string>()

const props = withDefaults(defineProps<{
  error?: string,
  inputWidth?: `${number}px`,
  label: string,
  labelPosition?: 'left' | 'right',
  placeholder?: string,
  shouldAutoselect?: boolean,
  type?: HTMLInputElement['type'],
}>(), {
  labelPosition: 'left',
})
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
    <label
      v-if="labelPosition === 'left'"
      :for="idInput"
    >
      {{ label }}
    </label>
    <InputText
      :id="idInput"
      v-model.trim="input"
      v-bind="$attrs"
      class="bc-input-text__input"
      :placeholder
      :type
    />
    <label
      v-if="labelPosition === 'right'"
      :for="idInput"
    >
      {{ label }}
    </label>
  </BcInputError>
</template>

<style lang="scss">
.bc-input-text__input {
  width: v-bind(inputWidth);
}
</style>
