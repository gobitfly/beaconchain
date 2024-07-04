<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { type MultiBarItem } from '~/types/multiBar'

interface Props {
  buttons: MultiBarItem[]
}

const props = defineProps<Props>()

const selected = defineModel<string[]>({ required: true })

const modelValues = ref<Record<string, boolean>>(props.buttons.reduce((map, { value }) => {
  map[value] = selected.value.includes(value)
  return map
}, {} as Record<string, boolean>))

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
    <BcToggleMultiBarButton
      v-for="button in props.buttons"
      :key="button.value"
      v-model="modelValues[button.value]"
      :class="button.className"
      :icon="button.icon"
      :tooltip="button.tooltip"
      :disabled="button.disabled"
    >
      <template #icon>
        <slot :name="button.value">
          <component :is="button.component" v-if="button.component" v-bind="button.componentProps" :class="button.componentClass" />
          <FontAwesomeIcon v-else-if="button.icon" :icon="button.icon" />
        </slot>
      </template>
    </BcToggleMultiBarButton>
  </div>
</template>

<style lang="scss" scoped>
.bc-togglebar {
  display: inline-flex;
  gap: var(--padding-small);
  padding: 7px 10px;

  background-color: var(--container-background);
  border: solid 1px var(--container-border-color);
  border-radius: var(--border-radius);
}
</style>
