<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { type MultiBarItem } from '~/types/multiBar'

interface Props {
  icons: MultiBarItem[]
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
    <BcToggleMultiBarButton
      v-for="icon in props.icons"
      :key="icon.value"
      v-model="modelValues[icon.value]"
      :class="icon.className"
      :icon="icon.icon"
      :tooltip="icon.tooltip"
    >
      <template #icon>
        <slot :name="icon.value">
          <component :is="icon.component" v-if="icon.component" />
          <FontAwesomeIcon v-else-if="icon.icon" :icon="icon.icon" />
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
