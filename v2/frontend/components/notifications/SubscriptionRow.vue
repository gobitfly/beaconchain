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
// const textField = ref('')

const { t } = useI18n()

interface CheckboxAndText {
  check: boolean,
  text: string
}

/** - if boolean: checked state of the checkbox
 *  - if number: -1 means unchecked, another value means checked and copies the textfield (converted in a number)
 *  - if array: list of selected networks */
const state = defineModel<boolean|number|ChainID[]>({ required: true })
let networkSelectorState: ModelRef<ChainID[]>
let checkboxAndText: Ref<CheckboxAndText> | undefined

// bridging the v-model with the data structures of this component

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
      output.check = (state as number >= 0)
      output.text = correctUserInput(output.check, output.check ? String(state) : checkboxAndText!.value.text /* textField.value */)
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
    case 'percent' :
      return state.check ? calculateCorrectNumber(state.text) : -1
    case 'binary' :
      return state.check
  }
  return 0
}

// input validation / autocorrection

function correctUserInput (checked: boolean, input: string) : string {
  switch (type.value) {
    case 'amount' : return (isThisAvalidInput(input) || checked) ? String(calculateCorrectNumber(input)) : ''
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

// calculation of the information to show in the tooltip

const tooltipLines = computed(() => {
  let options
  if (Array.isArray(state.value)) {
    options = { plural: state.value.length, list: state.value.join(', ') }
  } else {
    let plural: number
    if (type.value === 'amount' || type.value === 'percent') {
      plural = calculateCorrectNumber(checkboxAndText!.value.text)
    } else {
      plural = state.value ? 2 : 1
    }
    options = { plural }
  }
  return tAll(t, props.tPath + '.hint', options)
})

/*
// input: keeping the text field up-to-date with new data coming from the parent through the bridge
if (checkboxAndText) {
  watch(checkboxAndText, (cat) => {
    textField.value = cat.text
  }, { immediate: true, deep: true })
} */

// output: on blur or enter, auto-correcting the text field /*and sending the corrected data to the bridge*/

function acknowledgeInputtedText () {
  // textField.value = correctUserInput(checkboxAndText!.value.check, textField.value)
  // checkboxAndText!.value.text = textField.value
  checkboxAndText!.value.text = correctUserInput(checkboxAndText!.value.check, checkboxAndText!.value.text)
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
          v-if="checkboxAndText"
          v-model="checkboxAndText.text"
          type="text"
          :placeholder="t(tPath + '.placeholder')"
          :class="[deactivationClass,type]"
          @blur="acknowledgeInputtedText"
          @keypress.enter="acknowledgeInputtedText"
        />
        &nbsp;
      </div>
      <span v-if="type == 'percent'" :class="deactivationClass">%</span>
      <Checkbox
        v-model="checkboxAndText!.check"
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
