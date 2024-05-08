<script setup lang="ts">
interface Props {
  trueOption?: string,
  falseOption?: string,
}
const props = defineProps<Props>()
console.log(props) // TODO: Remove once done :()

const selected = defineModel<boolean>({ required: true })

</script>

<template>
  <div class="toggle-container">
    <slot name="falseOption">
      <div v-if="falseOption" class="option-label" :class="{ selected: !selected }">
        {{ falseOption }}
      </div>
    </slot>
    <InputSwitch v-model="selected" />
    <slot name="trueOption">
      <div v-if="trueOption" class="option-label" :class="{ selected }">
        {{ trueOption }}
      </div>
    </slot>
  </div>
</template>

<style lang="scss" scoped>
.toggle-container{
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
