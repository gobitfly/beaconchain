<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { InternalEntry } from '~/types/notifications/subscriptionModal'

const props = defineProps<{
  tPath: string,
  lacksPremiumSubscription: boolean
  valueInText?: number
}>()

const emitEvent = defineEmits<{(e: 'checkboxClick', checked: boolean) : void}>()

const { t } = useTranslation()

const parentVmodel = defineModel<InternalEntry>({ required: true })

const tooltipLines = computed(() => {
  let options
  if (props.valueInText !== undefined) {
    options = { plural: props.valueInText }
  } else
    if (Array.isArray(parentVmodel.value)) {
      options = { plural: parentVmodel.value.length, list: parentVmodel.value.join(', ') }
    } else {
      let plural: number
      if (parentVmodel.value.type === 'amount' || parentVmodel.value.type === 'percent') {
        plural = isNaN(parentVmodel.value.num!) ? 0 : parentVmodel.value.num!
      } else {
        plural = parentVmodel.value ? 2 : 1
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
    <BcTooltip v-if="tooltipLines[0]" :fit-content="true">
      <FontAwesomeIcon :icon="faInfoCircle" class="info" />
      <template #tooltip>
        <div v-if="tPath.includes('is_validator_offline_subscribed')" class="tt-content">
          {{ tooltipLines[0] }}
          <ul>
            <li>{{ tooltipLines[1] }}</li>
            <li>{{ tooltipLines[2] }}</li>
            <li>{{ tooltipLines[3] }}</li>
          </ul>
        </div>
        <div v-else-if="tPath.includes('group_offline_threshold')" class="tt-content">
          {{ tooltipLines[0] }}
          <ul>
            <li>{{ tooltipLines[1] }}</li>
            <li>{{ tooltipLines[2] }}</li>
            <li>{{ tooltipLines[3] }}</li>
            <li>{{ tooltipLines[4] }}</li>
          </ul>
          <b>{{ tooltipLines[5] }}</b> {{ tooltipLines[6] }}
        </div>
        <div v-else-if="tPath.includes('is_ignore_spam_transactions_enabled')" class="tt-content">
          {{ tooltipLines[0] }}
          <b>{{ tooltipLines[1] }}</b>
          {{ tooltipLines[2] }}
          <b>{{ tooltipLines[3] }}</b>{{ tooltipLines[4] }}
        </div>
        <div v-else class="tt-content">
          {{ tooltipLines[0] }}
        </div>
      </template>
    </BcTooltip>
    <BcPremiumGem v-if="lacksPremiumSubscription" class="gem" />
    <div v-if="parentVmodel.type != 'networks'" class="right">
      <div v-if="parentVmodel.type == 'amount' || parentVmodel.type == 'percent'" class="input">
        <BcInputNumber
          v-if="parentVmodel.num !== undefined"
          v-model="parentVmodel.num"
          :min="(parentVmodel.type === 'amount') ? 0 : 1"
          :max="(parentVmodel.type === 'amount') ? 2**32 : 100"
          :max-fraction-digits="(parentVmodel.type === 'amount') ? 2 : 1"
          :placeholder="t(tPath + '.placeholder')"
          :class="[deactivationClass,parentVmodel.type]"
        />
        &nbsp;
      </div>
      <span v-if="parentVmodel.type == 'percent'" :class="deactivationClass">%</span>
      <Checkbox
        v-if="parentVmodel.check !== undefined"
        v-model="parentVmodel.check"
        :binary="true"
        class="checkbox"
        :class="deactivationClass"
        @click="emitEvent('checkboxClick', !parentVmodel.check)"
      />
    </div>
    <div v-else class="right">
      <BcNetworkSelector v-if="parentVmodel.networks !== undefined" v-model="parentVmodel.networks" :class="deactivationClass" />
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
    padding-left: 1.4em;
  }
  .italic {
    font-style: italic;
  }
  .bold {
    font-weight: bold;
  }
}
</style>
