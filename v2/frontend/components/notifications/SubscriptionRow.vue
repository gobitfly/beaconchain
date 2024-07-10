<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ModelRef } from 'vue'
import type { ChainIDs } from '~/types/network'
import type { CheckboxAndNumber } from '~/types/notifications/subscriptionModal'

const props = defineProps<{
  tPath: string,
  lacksPremiumSubscription: boolean,
  inputType?: 'binary' | 'amount' | 'percent' | 'networks',
  default?: number
  valueInText?: number
}>()

const type = computed(() => props.inputType ?? 'binary')

const { t } = useI18n()

const state = defineModel<CheckboxAndNumber|ChainIDs[]>({ required: true })
let networkSelectorState: ModelRef<ChainIDs[]>
let checkBoxAndInput: Ref<CheckboxAndNumber>
if (type.value === 'networks') {
  networkSelectorState = state as ModelRef<ChainIDs[]>
} else {
  checkBoxAndInput = state as ModelRef<CheckboxAndNumber>
}

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
        plural = checkBoxAndInput!.value.num ?? Math.abs(props.default ?? 0)
      } else {
        plural = state.value ? 2 : 1
      }
      options = { plural }
    }
  return tAll(t, props.tPath + '.hint', options)
})

const deactivationClass = props.lacksPremiumSubscription ? 'deactivated' : ''
</script>

<template>
  <div class="option-row">
    <span class="caption" :class="deactivationClass">
      {{ t(tPath+'.option') }}
    </span>
    {{ checkBoxAndInput?.num }}
    <BcTooltip v-if="tooltipLines[0]" :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" class="info" />
      <template #tooltip>
        <BcMiniParser :input="tooltipLines" class="tt-content" />
      </template>
    </BcTooltip>
    <BcPremiumGem v-if="lacksPremiumSubscription" class="gem" />
    <div v-if="type != 'networks'" class="right">
      <div v-if="type == 'amount' || type == 'percent'" class="input">
        <BcInputNumber
          v-if="checkBoxAndInput"
          v-model="checkBoxAndInput.num"
          :min="(type === 'amount') ? 0 : 1"
          :max="(type === 'amount') ? 2**32 : 100"
          :max-fraction-digits="(type === 'amount') ? 2 : 1"
          :placeholder="t(tPath + '.placeholder')"
          :class="[deactivationClass,type]"
        />
        &nbsp;
      </div>
      <span v-if="type == 'percent'" :class="deactivationClass">%</span>
      <Checkbox
        v-model="checkBoxAndInput!.check"
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
    margin-right: var(--padding-small);
    &.deactivated {
      opacity: unset;
      color: var(--text-color-disabled);
    }
  }

  .info {
    margin-left: var(--padding-small);
    margin-right: var(--padding-small);
  }
  .gem {
    margin-left: var(--padding-small);
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
