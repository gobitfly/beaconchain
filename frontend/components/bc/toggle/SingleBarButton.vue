<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import BcTooltip from '../BcTooltip.vue'

interface Props {
  disabled?: boolean,
  icon?: IconDefinition,
  layout: 'gaudy' | 'minimal',
  selected: boolean,
  subText?: string,
  text?: string,
  tooltip?: string,
}
const props = defineProps<Props>()

const topBottomPadding = computed(() => (props.subText ? '8px' : '16px'))
</script>

<template>
  <BcTooltip
    :dont-open-permanently="true"
    :hover-delay="350"
    :hide="!tooltip"
  >
    <template #tooltip>
      <div class="button-tooltip">
        <div
          v-if="tooltip"
          class="individual"
        >
          {{ tooltip }}
        </div>
        <div>
          {{
            disabled
              ? $t("common.unavailable")
              : selected
                ? $t("common.selected")
                : $t("common.deselected")
          }}
        </div>
      </div>
    </template>
    <ToggleButton
      class="bc-toggle"
      :class="layout"
      :disabled="disabled"
      :model-value="selected"
    >
      <template #icon="slotProps">
        <slot
          name="icon"
          v-bind="slotProps"
        >
          <FontAwesomeIcon
            v-if="icon"
            :icon="icon"
          />
        </slot>
        <div
          v-if="text"
          class="label"
        >
          {{ text }}
          <div
            v-if="subText"
            class="sub"
          >
            {{ subText }}
          </div>
        </div>
      </template>
    </ToggleButton>
  </BcTooltip>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.button-tooltip {
  width: max-content;
  text-align: left;
  .individual {
    margin-bottom: var(--padding);
  }
}
.bc-toggle {
  min-width: 30px;
  min-height: 30px;
  &.p-button {
    &.p-togglebutton {
      &.gaudy {
        display: flex;
        flex-grow: 1;
        flex-direction: column;
        width: 100%;
        height: 100%;
        gap: 11px;
        padding: v-bind(topBottomPadding) 0;
        border: 1px var(--container-border-color) solid;
        border-radius: var(--border-radius);
        background-color: var(--container-background);
        color: var(--text-color);

        &.p-highlight {
          border-color: var(--button-color-active);
          color: var(--button-color-active);
        }
      }
      &.minimal {
        padding: 2px;
        border-style: none;
        color: var(--container-color);
        background-color: var(--container-border-color);

        &:not(.p-highlight) {
          background-color: var(--container-background);
        }
      }
      :deep(.p-button-label) {
        display: none;
      }

      :deep(svg) {
        max-width: 36px;
      }
      &.p-disabled {
        opacity: 0.5;
        cursor: default;
      }
    }
  }

  .label {
    @include fonts.subtitle_text;
    .sub {
      font-size: var(--tiny_text_font_size);
    }
  }
}
</style>
