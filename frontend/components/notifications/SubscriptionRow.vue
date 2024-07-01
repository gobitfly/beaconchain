<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { ChainIDs } from '~/types/network'

const props = defineProps<{
  tPath: string,
  lacksPremiumSubscription: boolean,
  inputType?: 'binary' | 'percent' | 'number' | 'networks'
}>()

const { t } = useI18n()

const type = computed(() => props.inputType || 'binary')
const liveState = defineModel<boolean|number|ChainIDs[]>({ required: true })
const checked = ref<boolean>(false)
const inputted = ref<string>('')

const tooltipLines = computed(() => {
  let options
  if (Array.isArray(liveState.value)) {
    options = { plural: liveState.value.length, list: liveState.value.join(', ') }
  } else {
    const plural = (type.value === 'number' || type.value === 'percent') ? liveState.value : (liveState.value ? 2 : 1)
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
        <BcMiniParser :input="tooltipLines" class="tt-content" />
      </template>
    </BcTooltip>
    <BcPremiumGem v-if="lacksPremiumSubscription" class="gem" />
    <div v-if="type != 'networks'" class="right">
      <InputText v-if="type == 'number' || type == 'percent'" v-model="inputted" :placeholder="t(tPath + '.placeholder')" :class="[deactivationClass,type]" />
      <span v-if="type == 'percent'" :class="deactivationClass">%</span>
      <Checkbox v-model="checked" :binary="true" class="checkbox" :class="deactivationClass" />
    </div>
    <div v-else class="right">
      <NotificationsNetworkSelector />
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
  @include fonts.small_text;
  height: 35px;
  align-items: center;

  .info {
    margin-left: 6px;
  }
  .gem {
    margin-left: 6px;
  }
  .right {
    display: flex;
    margin-left: auto;
    height: 100%;
    align-items: center;

    .number {
      width: 110px;
      margin-right: var(--padding-small);
    }
    .percent {
      width: 34px;
      margin-right: var(--padding-small);
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
