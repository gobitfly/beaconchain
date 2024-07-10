<script setup lang="ts">
import type { Nullable } from 'primevue/ts-helpers'
import InputNumber from 'primevue/inputnumber'

const props = defineProps<{
  min: number,
  max: number,
  maxFractionDigits: number
}>()

const parentVmodel = defineModel<number|null>({ required: true })

const bridgedVmodel = usePrimitiveRefBridge<number|null, Nullable<number>>(parentVmodel, n => n, n => n ?? null)

function sendValueIfValid (input: Nullable<number>) : void {
  if (input !== undefined && input !== null) {
    if (isNaN(input)) {
      input = null
    } else {
      if (input < props.min || input > props.max) {
        return
      }
      const stringifyied = String(input)
      const comma = stringifyied.indexOf('.')
      if (comma >= 0 && stringifyied.length - comma - 1 > props.maxFractionDigits) {
        return
      }
    }
  }
  bridgedVmodel.deactivateBridge() // this allows us to output to the parent v-model the value without causing an injection of the value into the InputNumber v-model (that would trigger the autocorrect of InputNumber at each key stroke)
  parentVmodel.value = input ?? null
  bridgedVmodel.reactivateBridge()
}
</script>

<template>
  <InputNumber
    v-model="bridgedVmodel"
    :min="min"
    :max="max"
    :max-fraction-digits="maxFractionDigits"
    class="why-the-hell-dont-they-fix-this-bug"
    @input="input => { if (typeof input.value !== 'string') sendValueIfValid(input.value) }"
  />
</template>

<style scoped lang="scss">
.why-the-hell-dont-they-fix-this-bug {
  :deep(input) { width: 100%; }
}
</style>
