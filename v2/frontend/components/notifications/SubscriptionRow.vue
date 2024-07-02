<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ModelRef } from 'vue'
import type { ChainID } from '~/types/network'

const props = defineProps<{
  tPath: string,
  lacksPremiumSubscription: boolean,
  inputType?: 'binary' | 'amount' | 'percent' | 'networks',
  default?: number
}>()

const { t } = useI18n()
const { bridgeArraysRefs } = useRefBridge()

const type = computed(() => props.inputType || 'binary')
const state = defineModel<boolean|number|ChainID[]>({ required: true })
const networkSelectorState = ref<ChainID[]>([])
const checked = ref<boolean>(false)
const inputted = ref('')

refreshUIfromState() // initial loading

if (props.inputType === 'networks') {
  bridgeArraysRefs(state as ModelRef<ChainID[]>, networkSelectorState)
}

const tooltipLines = computed(() => {
  let options
  if (Array.isArray(state.value)) {
    options = { plural: state.value.length, list: state.value.join(', ') }
  } else {
    let plural: number
    if (type.value === 'amount' || type.value === 'percent') {
      plural = calculateCorrectNumber(inputted)
    } else {
      plural = state.value ? 2 : 1
    }
    options = { plural }
  }
  return tAll(t, props.tPath + '.hint', options)
})

function outputSetting () : void {
  // outputs the setting
  switch (type.value) {
    case 'amount' :
      state.value = checked.value && isThisAvalidInput(inputted) ? calculateCorrectNumber(inputted) : -1
      break
    case 'percent' :
      state.value = checked.value ? calculateCorrectNumber(inputted) : -1
      break
    case 'binary' :
      state.value = checked.value
      break
  }
}

function correctUserInput () : void {
  switch (type.value) {
    case 'amount' :
      inputted.value = isThisAvalidInput(inputted) ? String(calculateCorrectNumber(inputted)) : ''
      break
    case 'percent' :
      inputted.value = String(calculateCorrectNumber(inputted))
      break
  }
}

function calculateCorrectNumber (input: string | Ref<string>) : number {
  if (typeof input !== 'string') {
    input = input.value
  }
  let num = !isThisAvalidInput(input) ? (props.default ?? 0) : Number(input)
  if (type.value === 'percent') {
    if (num < 1) { num = 1 }
    if (num > 100) { num = 100 }
  }
  return num
}

function isThisAvalidInput (input: string | Ref<string>) : boolean {
  if (typeof input !== 'string') {
    input = input.value
  }
  return !!input && !isNaN(Number(input)) && Number(input) >= 0
}

function refreshUIfromState () : void {
  switch (type.value) {
    case 'amount' :
    case 'percent' :
      state.value = state.value as number
      inputted.value = String(state.value)
      checked.value = (state.value >= 0)
      correctUserInput()
      break
    case 'binary' :
      checked.value = state.value as boolean
      break
    case 'networks' :
      break
  }
}

const deactivationClass = props.lacksPremiumSubscription ? 'deactivated' : ''
</script>

<template>
  <div class="option-row">
    <span class="caption" :class="deactivationClass">
      {{ t(tPath+'.option') }}
    </span>
    <BcTooltip v-if="tooltipLines[0]" :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" class="info" />
      <template #tooltip>
        <BcMiniParser :input="tooltipLines" class="tt-content" />
      </template>
    </BcTooltip>
    <BcPremiumGem v-if="lacksPremiumSubscription" />
    <div v-if="type != 'networks'" class="right">
      <div v-if="type == 'amount' || type == 'percent'" class="input">
        <InputText
          v-model="inputted"
          type="text"
          :placeholder="t(tPath + '.placeholder')"
          :class="[deactivationClass,type]"
          @change="outputSetting"
          @blur="correctUserInput"
        />
        &nbsp;
      </div>
      <span v-if="type == 'percent'" :class="deactivationClass">%</span>
      <Checkbox
        v-model="checked"
        :binary="true"
        class="checkbox"
        :class="deactivationClass"
        @change="outputSetting"
      />
    </div>
    <div v-else class="right">
      <BcNetworkSelector v-model="networkSelectorState" />
    </div>
  </div>
</template>

<style scoped lang="scss">
@use "~/assets/css/fonts.scss";

.deactivated {
  opacity: 0.6;
  pointer-events: none;
}

.option-row {
  display: flex;
  position: relative;
  @include fonts.small_text;
  align-items: center;

  .info {
    margin-left: 6px;
    margin-right: 6px;
  }
  .right {
    display: flex;
    position: relative;
    margin-left: auto;
    height: 100%;
    align-items: center;

    .input {
      position: relative;
      height: 100%;
      width: 100%;
      .amount,
      .percent {
        position: absolute;
        height: 29px;
        right: 0px;
        top: -6px;
        margin-right: var(--padding-small);
      }
      .amount {
        width: 110px;
      }
      .percent {
        width: 43px;
      }
    }
    .checkbox {
      margin-left: var(--padding-small);
    }
  }
}

.tt-content {
  width: 220px;
  min-width: 100%;
  text-align: left;
  ul {
    padding: 0;
    margin: 0;
    padding-left: 1.3em;
    li::marker {
      font-size: 0.6rem;
    }
  }
  .italic {
    font-style: italic;
  }
  .bold {
    font-weight: bold;
  }
}
</style>
