<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { type MultiBarItem } from '~/types/multiBar'

interface Props {
  displayMode?: boolean,
  buttons: MultiBarItem[]
}

const props = defineProps<Props>()

type ButtonStates = Record<string, boolean>

const inOutSelection = defineModel<string[]>({ required: true })
const buttonStates = useObjectRefBridge<string[], ButtonStates>(inOutSelection, receiveFromVModel, sendToVModel, true)

function receiveFromVModel (selection: string[]) : ButtonStates {
  const states = props.buttons.reduce((map, { value }) => {
    map[value] = selection.includes(value)
    return map
  }, {} as ButtonStates)
  return states
}

function sendToVModel (states: ButtonStates) : string[] {
  const selection: string[] = []
  Object.entries(states).forEach(([key, value]) => {
    if (value) {
      selection.push(key)
    }
  })
  return selection
}

// this line is independent of the bridge above (that addresses the on/off states), this line updates the component if the list of buttons comes late
watch(() => props.buttons, () => { buttonStates.value = receiveFromVModel(inOutSelection.value) })

const displayModeClass = computed(() => props.displayMode ? 'read-only' : '')
</script>

<template>
  <div class="bc-togglebar" :class="displayModeClass">
    <BcToggleMultiBarButton
      v-for="button in props.buttons"
      :key="button.value"
      v-model="buttonStates[button.value]"
      :class="button.className"
      :icon="button.icon"
      :tooltip="button.tooltip"
      :disabled="button.disabled"
      :display-mode-class="displayModeClass"
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

  &.read-only {
    padding: 0px;
    border: none;
  }
}
</style>
