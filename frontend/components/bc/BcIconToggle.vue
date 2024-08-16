<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

interface Props {
  disabled?: boolean,
  falseIcon?: IconDefinition,
  trueIcon?: IconDefinition,
}

const props = defineProps<Props>()

const selected = defineModel<boolean>({ required: true })

const toggle = () => {
  if (props.disabled) {
    return
  }
  if (selected.value === undefined) {
    return
  }
  selected.value = !selected.value
}
</script>

<template>
  <div
    class="bc-toggle"
    :class="{ selected }"
    :disabled="disabled || null"
    @click="toggle"
  >
    <div class="icon true-icon">
      <slot name="trueIcon">
        <FontAwesomeIcon
          v-if="trueIcon"
          :icon="trueIcon"
        />
      </slot>
    </div>
    <div class="icon false-icon">
      <slot name="falseIcon">
        <FontAwesomeIcon
          v-if="falseIcon"
          :icon="falseIcon"
        />
      </slot>
    </div>
    <span class="slider" />
    <div class="bg" />
  </div>
</template>

<style lang="scss" scoped>
.bc-toggle {
  position: relative;
  width: 54px;
  height: 23px;
  cursor: pointer;

  &[disabled] {
    cursor: unset;
    pointer-events: none;
    opacity: 0.5;
  }

  .bg {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border-radius: var(--border-radius);
    background-color: var(--light-grey-2);
    border: 1px solid var(--light-grey-3);
    z-index: 1;
  }

  .slider {
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    width: 50%;
    transform: translateX(100%);
    transition: 0.2s transform;
    background-color: var(--primary-color);
    border-radius: var(--border-radius);
    z-index: 2;
  }

  .icon {
    width: 50%;
    height: 100%;
    display: inline-flex;
    justify-content: center;
    align-items: center;
    color: var(--text-color);
    transition: 0.2s color;
    position: relative;
    z-index: 3;

    :deep(svg) {
      max-width: 18px;
      max-height: 18px;
    }
  }

  &.selected {
    .slider {
      transform: translateX(0);
    }

    .icon.true-icon {
      color: var(--text-color-inverted);
    }
  }

  &:not(.selected) {
    .icon.false-icon {
      color: var(--text-color-inverted);
    }
  }
}

.dark-mode .bc-toggle .bg {
  background-color: var(--graphite);
  border-color: var(--graphite);
}
</style>
