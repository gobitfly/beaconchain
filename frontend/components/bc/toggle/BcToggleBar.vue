<script setup lang="ts">
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import type { Component } from 'vue'

interface Props {
  icons: {
    icon?: IconDefinition
    component?: Component,
    value: string
  }[]
}

const props = defineProps<Props>()

const selected = defineModel<string[]>({ required: true })

const inital: Record<string, boolean> = {}
const modelValues = ref<Record<string, boolean>>(props.icons.reduce((map, { value }) => {
  map[value] = selected.value.includes(value)
  return map
}, inital))

watch(modelValues, () => {
  const list: string[] = []
  Object.entries(modelValues.value).forEach(([key, value]) => {
    if (value) {
      list.push(key)
    }
  })
  selected.value = list
}, { deep: true })

</script>

<template>
  <div class="bc-togglebar">
    <BcToggleButton v-for="icon in props.icons" :key="icon.value" v-model="modelValues[icon.value]" :icon="icon.icon">
      <template #icon>
        <slot :name="icon.value">
          <component :is="icon.component" />
        </slot>
      </template>
    </BcToggleButton>
  </div>
</template>

<style lang="scss" scoped>
.bc-togglebar {
  display: inline-flex;
  gap: var(--padding);
  padding: 7px 10px;

  background-color: var(--container-background);
  border: solid 1px var(--container-border-color);
  border-radius: var(--border-radius);
}
</style>
