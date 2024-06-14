<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import type { Component } from 'vue'
interface Props {
  buttons: {
    icon?: IconDefinition,
    text?: string,
    subText?: string,
    component?: Component,
    componentProps?: any,
    componentClass?: string,
    value: string,
    disabled?: boolean
  }[],
  allowDeselect?: boolean // if true, clicking the selected button will deselect it causing the whole SingleBar not to have a value
}
const props = defineProps<Props>()

const selected = defineModel<string>()

const values = ref<Record<string, boolean>>(props.buttons.reduce((map, { value }) => {
  map[value] = value === selected.value
  return map
}, {} as Record<string, boolean>))

function onButtonClicked (value: string) {
  for (const key in values.value) {
    if (key === value) {
      if (values.value[key] && !props.allowDeselect) {
        continue
      }
      values.value[key] = !values.value[key]
    } else {
      values.value[key] = false
    }
  }
  selected.value = values.value[value] ? value : ''
}
</script>

<template>
  <div class="bc-togglebar">
    <BcToggleSingleBarButton
      v-for="button in props.buttons"
      :key="button.value"
      :icon="button.icon"
      :text="button.text"
      :sub-text="button.subText"
      :selected="values[button.value]"
      :disabled="button.disabled"
      @click="!button.disabled && onButtonClicked(button.value)"
    >
      <template #icon>
        <slot :name="button.value">
          <component :is="button.component" v-bind="button.componentProps" :class="button.componentClass" />
        </slot>
      </template>
    </BcToggleSingleBarButton>
  </div>
</template>

<style lang="scss" scoped>
.bc-togglebar {
  display: inline-flex;
  gap: var(--padding);
}
</style>
