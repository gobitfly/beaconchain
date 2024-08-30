<script setup lang="ts">
interface Props {
  disabled?: boolean,
  falseOption?: string,
  trueOption?: string,
}
defineProps<Props>()

const selected = defineModel<boolean>({ required: true })
</script>

<template>
  <div class="toggle-container">
    <slot name="falseOption">
      <div
        v-if="falseOption"
        class="option-label"
        :class="{ selected: !selected }"
      >
        {{ falseOption }}
      </div>
    </slot>
    <ToggleSwitch
      v-model="selected"
      :disabled
    />
    <slot name="trueOption">
      <div
        v-if="trueOption"
        class="option-label"
        :class="{ selected }"
      >
        {{ trueOption }}
      </div>
    </slot>
  </div>
</template>

<style lang="scss" scoped>
.toggle-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 13px;

  .option-label {
    &.selected {
      color: var(--text-color);
    }

    &:not(.selected) {
      color: var(--text-color-discreet);
    }
  }
}
</style>
