<script setup lang="ts">
import BcTooltip from '../BcTooltip.vue'
import type { Icon } from '~/components/bc/icon/BcIcon.vue'

interface Props {
  disabled?: boolean,
  falseIcon?: Icon,
  icon?: Icon,
  readonlyClass?: string,
  tooltip?: string,
}

const props = defineProps<Props>()

const selected = defineModel<boolean | undefined>({ required: true })

const icon = computed(() => {
  return selected.value || !props.falseIcon ? props.icon : props.falseIcon
})
</script>

<template>
  <BcTooltip
    :dont-open-permanently="true"
    :hover-delay="350"
  >
    <template #tooltip>
      <div
        class="button-tooltip"
        :class="readonlyClass"
      >
        <div
          v-if="tooltip"
          class="individual"
        >
          {{ tooltip }}
        </div>
        <div v-if="readonlyClass !== 'read-only'">
          {{
            disabled
              ? $t("common.unavailable")
              : selected
                ? $t("filter.enabled")
                : $t("filter.disabled")
          }}
        </div>
      </div>
    </template>
    <ToggleButton
      v-model="selected"
      class="bc-toggle"
      :class="readonlyClass"
      :disabled="disabled || readonlyClass === 'read-only'"
    >
      <template #icon="slotProps">
        <slot
          name="icon"
          v-bind="slotProps"
        >
          <BcIcon
            v-if="icon"
            :name="icon"
          />
        </slot>
      </template>
    </ToggleButton>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.button-tooltip {
  width: max-content;
  text-align: left;

  .individual::not(.read-only) {
    margin-bottom: var(--padding);
  }
}

.bc-toggle {
  min-width: 30px;
  min-height: 30px;

  &.p-togglebutton {
    padding: 2px;
    border-style: none;
    color: var(--container-color);
    background-color: var(--container-border-color);
    border-radius: var(--border-radius);

    &:not(.p-togglebutton-checked),
    &.read-only {
      background-color: var(--container-background);
    }

    :deep(.p-togglebutton-content) {
      width: 24px;
      height: 24px;
    }

    // this is needed as the primvevue ToggleButton adds a yes/no label if none is provided
    :deep(.p-togglebutton-label) {
      display: none;
    }

    &.p-disabled {
      cursor: default;

      &:not(.read-only) {
        opacity: 0.5;
      }
    }
  }
}
</style>
