<script setup lang="ts">
import type { Nullable } from 'primevue/ts-helpers'
import InputNumber from 'primevue/inputnumber'

const props = defineProps<{
  min: number
  max: number
  maxFractionDigits: number
}>()

const parentVmodel = defineModel<number>({ required: true })
const bridgedVmodel = usePrimitiveRefBridge<number, Nullable<number>>(parentVmodel, n => (isNaN(n) ? null : n), n => (n ?? NaN))

function sendValue(input: Nullable<number>): void {
  if (input === undefined || input === null || isNaN(input) || input < props.min || input > props.max) {
    input = NaN
  }
  else {
    const stringifyied = String(input)
    const comma = stringifyied.indexOf('.')
    if (comma >= 0 && stringifyied.length - comma - 1 > props.maxFractionDigits) {
      input = NaN
    }
  }
  bridgedVmodel.pauseBridgeFromNowOn() // this allows us to output the value to the parent v-model without causing an injection of the value back into the InputNumber v-model (that would empty InputNumber at each key stroke if the input is invalid)
  parentVmodel.value = input
  bridgedVmodel.wakeupBridgeAtNextTick()
}
</script>

<template>
  <InputNumber
    v-model="bridgedVmodel"
    :min="min"
    :max="max"
    :max-fraction-digits="maxFractionDigits"
    locale="en-US"
    class="why-the-hell-dont-they-fix-this-bug"
    @input="input => { if (typeof input.value !== 'string') sendValue(input.value) }"
  />
</template>

<style scoped lang="scss">
.why-the-hell-dont-they-fix-this-bug {
  :deep(input) { width: 100%; }
}
</style>
