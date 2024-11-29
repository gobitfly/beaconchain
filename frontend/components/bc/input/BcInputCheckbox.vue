<script lang="ts" setup>
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

const props = defineProps<{
  error?: string,
  infoText?: string,
  label?: string,
}>()

const id = props.label ? useId() : undefined
const input = defineModel<boolean>()
</script>

<template>
  <BcInputError :error>
    <span v-if="label">
      <label
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
      class="bc-input-ckeckbox__checkbox"
      :input-id="id"
      v-bind="$attrs"
      binary
    />
  </BcInputError>
</template>

<style lang="scss">
.label {
  cursor: pointer;
}
.bc-input-checkbox__info {
  margin-left: var(--padding);
}
.bc-input-ckeckbox__checkbox {
  --outline-width: 0.125rem;
  --outline-offset: 0.125rem;
  margin: calc(var(--outline-width) + var(--outline-offset));
}
.bc-input-ckeckbox__checkbox:has(input:focus-visible) {
  outline: var(--outline-width) solid var(--blue-500);
  outline-offset: var(--outline-offset);
  border-radius: 0.125rem;

}
</style>
