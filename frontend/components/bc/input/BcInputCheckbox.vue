<script lang="ts" setup>
import {
  faInfoCircle,
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { BcInputError } from '~/components/bc/input/BcInputError.vue'

const props = defineProps<{
  error?: BcInputError
  label?: string
  infoText?: string
}>()

const id = props.label ? useId() : undefined
const input = defineModel<boolean>()
</script>

<template>
  <BcInputError :error>
    <span>
      <label
        v-if="label"
        class="label"
        :for="id"
      >
        {{ label }}
      </label>
      <BcTooltip
        v-if="infoText || $slots.tooltip"
        class="bc-input-checkbox__info"
        tooltip-width="220px"
        tooltip-text-align="left"
      >
        <FontAwesomeIcon :icon="faInfoCircle" />
        <template #tooltip>
          {{ infoText }}
          <slot name="tooltip" />
        </template>
      </BcTooltip>
    </span>
    <Checkbox
      v-model="input"
      :input-id="id"
      v-bind="$attrs"
      binary
    />
  </BcInputError>
</template>

<style lang="scss">
.label{
  cursor: pointer;
}
.bc-input-checkbox__info {
  margin-left: var(--padding);
}
</style>
