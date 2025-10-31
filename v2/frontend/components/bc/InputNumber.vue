<script setup lang="ts">
import InputNumber from 'primevue/inputnumber'

const props = defineProps<{
  max: number,
  maxFractionDigits: number,
  min: number,
}>()

const parentVmodel = defineModel<number>({ required: true })
const bridgedVmodel = usePrimitiveRefBridge<number, null | number>(
  parentVmodel,
  n => (isNaN(n) ? null : n),
  n => n ?? NaN,
)

function sendValue(input?: null | number): void {
  if (
    input === undefined
    || input === null
    || isNaN(input)
    || input < props.min
    || input > props.max
  ) {
    input = NaN
  }
  else {
    const stringifyied = String(input)
    const comma = stringifyied.indexOf('.')
    if (
      comma >= 0
      && stringifyied.length - comma - 1 > props.maxFractionDigits
    ) {
      input = NaN
    }
  }
  // this allows us to output the value to the parent v-model without causing
  // an injection of the value back into the InputNumber v-model (that would
  // empty InputNumber at each key stroke if the input is invalid)
  bridgedVmodel.pauseBridgeFromNowOn()
  parentVmodel.value = input
  bridgedVmodel.wakeupBridgeAtNextTick()
}
</script>

<template>
  <InputNumber
    v-model="bridgedVmodel"
    :min
    :max
    :max-fraction-digits
    locale="en-US"
    class="why-the-hell-dont-they-fix-this-bug"
    @input="
      (input) => {
        if (typeof input.value !== 'string') sendValue(input.value);
      }
    "
  />
</template>

<style scoped lang="scss">
.why-the-hell-dont-they-fix-this-bug {
  :deep(input) {
    width: 100%;
  }
}
</style>
