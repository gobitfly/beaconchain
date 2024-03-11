<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

interface Props {
  icon?: IconDefinition,
  falseIcon?: IconDefinition,
}

const props = defineProps<Props>()

const selected = defineModel<boolean | undefined>({ required: true })

const icon = computed(() => {
  return selected.value || !props.falseIcon ? props.icon : props.falseIcon
})
</script>

<template>
  <ToggleButton v-model="selected" class="bc-toggle" on-label="" off-icon="">
    <template #icon="slotProps">
      <slot name="icon" v-bind="slotProps">
        <FontAwesomeIcon v-if="icon" :icon="icon" />
      </slot>
    </template>
  </ToggleButton>
</template>

<style lang="scss" scoped>
.bc-toggle {
  &.p-button {
    &.p-togglebutton {

      width: 30px;
      height: 30px;
      padding: 2px;
      border-style: none;

      &:not(.p-highlight) {
        background-color: var(--container-background);
        color: var(--container-color);
      }

      // this is needed as the primvevue ToggleButton adds a yes/no label if none is provided
      :deep(.p-button-label) {
        display: none;
      }
    }

  }
}
</style>
