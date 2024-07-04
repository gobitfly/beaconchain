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

const type = computed(() => props.inputType || 'binary')

const { t } = useI18n()

interface CheckboxAndText {
  check: boolean,
  text: string
}

const state = defineModel<boolean|number|ChainID[]>({ required: true })
let networkSelectorState: ModelRef<ChainID[]>
let checkboxAndText = undefined as unknown as Ref<CheckboxAndText>

if (type.value === 'networks') {
  networkSelectorState = state as ModelRef<ChainID[]>
} else {
  checkboxAndText = useObjectRefBridge<boolean|number, CheckboxAndText>(state as Ref<boolean|number>, receiveFromVModel, sendToVModel)
}

function receiveFromVModel (state: boolean|number) : CheckboxAndText {
  const output = {} as CheckboxAndText
  switch (type.value) {
    case 'amount' :
    case 'percent' :
      output.text = correctUserInput(String(state))
      output.check = (state as number >= 0)
      break
    case 'binary' :
      output.check = state as boolean
      break
  }
  return output
}

function sendToVModel (state: CheckboxAndText) : boolean|number {
  switch (type.value) {
    case 'amount' :
      return state.check && isThisAvalidInput(state.text) ? calculateCorrectNumber(state.text) : -1
    case 'percent' :
      return state.check ? calculateCorrectNumber(state.text) : -1
    case 'binary' :
      return state.check
  }
  return 0
}

function correctUserInput (input: string) : string {
  switch (type.value) {
    case 'amount' : return isThisAvalidInput(input) ? String(calculateCorrectNumber(input)) : ''
    case 'percent' : return String(calculateCorrectNumber(input))
  }
  return ''
}

function calculateCorrectNumber (input: string) : number {
  let num = !isThisAvalidInput(input) ? (props.default ?? 0) : Number(input)
  if (type.value === 'percent') {
    if (num < 1) { num = 1 }
    if (num > 100) { num = 100 }
  }
  return num
}

function isThisAvalidInput (input: string) : boolean {
  return !!input && !isNaN(Number(input)) && Number(input) >= 0
}

const tooltipLines = computed(() => {
  let options
  if (Array.isArray(state.value)) {
    options = { plural: state.value.length, list: state.value.join(', ') }
  } else {
    let plural: number
    if (type.value === 'amount' || type.value === 'percent') {
      plural = calculateCorrectNumber(checkboxAndText.value.text)
    } else {
      plural = state.value ? 2 : 1
    }
    options = { plural }
  }
  return tAll(t, props.tPath + '.hint', options)
})

const textField = ref('')

if (checkboxAndText) {
  watch(() => checkboxAndText.value.text, () => { textField.value = checkboxAndText.value.text }, { immediate: true })
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
          v-model="textField"
          type="text"
          :placeholder="t(tPath + '.placeholder')"
          :class="[deactivationClass,type]"
          @blur="checkboxAndText.text = correctUserInput(textField)"
        />
        &nbsp;
      </div>
      <span v-if="type == 'percent'" :class="deactivationClass">%</span>
      <Checkbox
        v-model="checkboxAndText.check"
        :binary="true"
        class="checkbox"
        :class="deactivationClass"
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
