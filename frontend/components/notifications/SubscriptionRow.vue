<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ModelRef } from 'vue'
import type { ChainIDs } from '~/types/network'
import type { CheckboxAndNumber } from '~/types/subscriptionModal'

const props = defineProps<{
  tPath: string,
  lacksPremiumSubscription: boolean,
  inputType?: 'binary' | 'amount' | 'percent' | 'networks',
  default?: number
  valueInText?: number
}>()

const type = computed(() => props.inputType ?? 'binary')

const { t } = useI18n()

interface CheckboxAndText {
  check: boolean,
  text: string
}

const state = defineModel<CheckboxAndNumber|ChainIDs[]>({ required: true })
let networkSelectorState: ModelRef<ChainIDs[]>
let checkboxAndText: Ref<CheckboxAndText> | undefined

// bridging the v-model (CheckboxAndNumber | ChainIDs[]) with the data of this component (CheckboxAndText | ChainIDs[])

if (type.value === 'networks') {
  networkSelectorState = state as ModelRef<ChainIDs[]>
} else {
  checkboxAndText = useObjectRefBridge<CheckboxAndNumber, CheckboxAndText>(state as Ref<CheckboxAndNumber>, receiveFromVModel, sendToVModel)
}

function receiveFromVModel (state: CheckboxAndNumber) : CheckboxAndText {
  const output = {} as CheckboxAndText
  output.check = state.check
  if (type.value === 'amount' || type.value === 'percent') {
    output.text = isNaN(state.num) ? '' : String(state.num)
  }
  return output
}

function sendToVModel (state: CheckboxAndText) : CheckboxAndNumber {
  const output = {} as CheckboxAndNumber
  output.check = state.check
  if (type.value === 'amount' || type.value === 'percent') {
    const corrected = correctUserInput(state.text)
    output.num = (corrected === '') ? NaN : Number(corrected)
  }
  return output
}

// input validation / autocorrection

function correctUserInput (input: string) : string {
  switch (type.value) {
    case 'amount' : return isThisAvalidInput(input) ? String(calculateCorrectNumber(input)) : ''
    case 'percent' : return String(calculateCorrectNumber(input))
  }
  return ''
}

function calculateCorrectNumber (input: string) : number {
  let num = !isThisAvalidInput(input) ? Math.abs(props.default ?? 0) : Number(input)
  if (type.value === 'percent') {
    if (num < 1) { num = 1 }
    if (num > 100) { num = 100 }
    num = Math.round(10 * num) / 10
  }
  return num
}

function isThisAvalidInput (input: string) : boolean {
  return !!input && !isNaN(Number(input)) && Number(input) >= 0
}

// calculation of the information to show in the tooltip

const tooltipLines = computed(() => {
  let options
  if (props.valueInText !== undefined) {
    options = { plural: props.valueInText }
  } else
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

// output: on blur or enter, auto-correcting the text field
function acknowledgeInputtedText () {
  checkboxAndText!.value.text = correctUserInput(checkboxAndText!.value.text)
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
    <BcPremiumGem v-if="lacksPremiumSubscription" class="gem" />
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
      <BcNetworkSelector v-model="networkSelectorState" :class="deactivationClass" />
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

  .caption {
    margin-right: 3px;
  }

  .info {
    margin-left: 3px;
    margin-right: 3px;
  }
  .gem {
    margin-left: 3px;
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
        width: 48px;
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
