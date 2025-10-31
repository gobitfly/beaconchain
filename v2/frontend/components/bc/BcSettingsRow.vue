<script setup lang="ts">
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

defineProps<{
  hasBorderTop?: boolean,
  hasPremiumGem?: boolean,
  hasUnit?: boolean,
  info?: string,
  label: string,
}>()
const idCheckbox = useId()
const checkbox = defineModel<boolean>('checkbox')
const input = defineModel<string>('input')
</script>

<template>
  <div
    class="bc-settings-row"
    :class="{ 'has-border-top': hasBorderTop }"
  >
    <span class="bc-settings-row--info">
      <label
        :for="idCheckbox"
      >
        {{ label }}
      </label>
      <BcTooltip
        v-if="info || $slots.info"
        tooltip-width="220px"
        tooltip-text-align="left"
      >
        <FontAwesomeIcon :icon="faInfoCircle" />
        <template #tooltip>
          <slot name="info">
            {{ info }}
          </slot>
        </template>
      </BcTooltip>
      <BcPremiumGem
        v-if="hasPremiumGem"
      />
    </span>
    <span class="bc-settings-row--action">
      <LazyBcInputUnit
        v-if="hasUnit"
        v-model="input"
        unit=" %"
        :disabled="hasPremiumGem"
      />
      <BcInputCheckbox
        v-model="checkbox"
        :input-id="idCheckbox"
        :disabled="hasPremiumGem"
      />
    </span>
  </div>
</template>

<style scoped lang="scss">
.bc-settings-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.bc-settings-row--info {
  display: flex;
  gap: 0.75rem;
  align-items: center
}
.bc-settings-row--action {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.has-border-top {
  border-top: 1px solid var(--dark-grey);
  padding-top: 0.5rem;
}
</style>
