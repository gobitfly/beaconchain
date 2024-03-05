<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import type { Component } from 'vue'
interface Props {
  buttons: {
    icon?: IconDefinition,
    text?: string,
    component?: Component,
    value: string
  }[],
  initial?: string
}
const props = defineProps<Props>()

const selected = defineModel<string>({ required: true })
selected.value = props.initial || ''

const modelValues = ref<Record<string, boolean>>(props.buttons.reduce((map, { value }) => {
  map[value] = value === props.initial
  return map
}, {} as Record<string, boolean>))

function onButtonClicked (value: string) {
  for (const key in modelValues.value) {
    if (key !== value) {
      modelValues.value[key] = false
    }
  }
  selected.value = modelValues.value[value] ? value : ''
}
</script>

<template>
  <div class="bc-togglebar">
    <BcToggleSingleButton
      v-for="button in props.buttons"
      :key="button.value"
      v-model="modelValues[button.value]"
      :icon="button.icon"
      :text="button.text"
      @click="onButtonClicked(button.value)"
    >
      <template #icon>
        <slot :name="button.value">
          <component :is="button.component" />
        </slot>
      </template>
    </BcToggleSingleButton>
  </div>
</template>

<style lang="scss" scoped>
.bc-togglebar {
  display: inline-flex;
  gap: var(--padding);
}
</style>
