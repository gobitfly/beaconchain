<script setup lang="ts">
import { useTooltipStore } from '~/stores/useTooltipStore'

interface Props {
  text?: string,
  title?: string,
  layout?: 'dark' | 'default'
  position?: 'top' | 'left' | 'right' | 'bottom',
  hide?: boolean
}

const props = defineProps<Props>()
const bcTooltip = ref<HTMLElement | null>(null)
const ttStore = useTooltipStore()
const { doSelect } = ttStore
const { selected } = storeToRefs(ttStore)

// this const will be avaiable on template
const slots = useSlots()

const hasContent = computed(() => !!slots.tooltip || !!props.text)
const canBeOpened = computed(() => !props.hide && hasContent.value)

const hover = ref(false)
const isSelected = computed(() => !!bcTooltip.value && selected.value === bcTooltip.value)
const isOpen = computed(() => isSelected.value || hover.value)

const pos = ref<{ top: string, left: string }>({ top: '0', left: '0' })

const classList = computed(() => {
  return [props.layout || 'default', props.position || 'bottom', isOpen.value ? 'open' : 'closed']
})

const setPosition = () => {
  const rect = bcTooltip.value?.getBoundingClientRect()
  if (!rect) {
    return
  }
  const padding = 4
  let top = rect.bottom + padding
  let left = rect.left + rect.width / 2
  switch (props.position) {
    case 'left':
      left = rect.left - padding
      top = rect.top + rect.height / 2
      break
    case 'top':
      top = rect.top - padding
      break
    case 'right':
      left = rect.left + rect.width + padding
      top = rect.top + rect.height / 2
      break
  }
  pos.value = { top: `${top}px`, left: `${left}px` }
}

const handleClick = () => {
  if (isSelected.value) {
    doSelect(null)
  } else if (canBeOpened.value) {
    doSelect(bcTooltip.value)
    setPosition()
  }
}

const onHover = (enter: boolean) => {
  if (!enter) {
    hover.value = false
  } else if (canBeOpened.value && !selected.value) {
    hover.value = true
    setPosition()
  }
}

const doHide = (event: MouseEvent) => {
  if (event.target === bcTooltip.value || isParent(bcTooltip.value, event.target as HTMLElement)) {
    return
  }
  if (isSelected.value) {
    doSelect(null)
  }
  hover.value = false
  if (!isOpen.value) {
    bcTooltip.value?.blur()
  }
}

onMounted(() => {
  document.addEventListener('click', doHide)
})

onUnmounted(() => {
  document.removeEventListener('click', doHide)
  if (isSelected.value) {
    doSelect(null)
  }
})

</script>
<template>
  <div
    ref="bcTooltip"
    class="slot_container"
    @mouseover="onHover(true)"
    @mouseleave="hover = false"
    @click="handleClick()"
    @blur="onHover(false)"
  >
    <slot />
    <Teleport v-if="isOpen" to="body">
      <div class="bc-tooltip-wrapper" :style="pos">
        <div class="bc-tooltip" :class="classList" @click="$event.stopImmediatePropagation()">
          <slot name="tooltip">
            <b v-if="props.title">
              {{ props.title }}
            </b>
            {{ text }}
          </slot>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style lang="scss" scoped>
.slot_container {
  display: inline;

  &.active {
    cursor: pointer;
  }
}

.bc-tooltip-wrapper {
  position: fixed;
  width: 1px;
  height: 1px;
  overflow: visible;
  z-index: 99999;

}

.bc-tooltip {

  --tt-bg-color: var(--light-grey-2);
  --tt-color: var(--light-black);

  position: relative;
  display: inline-flex;
  flex-wrap: wrap;
  opacity: 0;
  transition: opacity 1s;
  text-align: center;
  padding: 9px 12px;
  min-width: 120px;
  border-radius: var(--border-radius);
  color: var(--tt-color);
  background: var(--tt-bg-color);
  font-family: var(--inter-family);
  font-weight: var(--inter-light);
  font-size: 10px;
  pointer-events: none;
  transform: translate(-50%, 0);

  &.dark {
    --tt-bg-color: var(--light-black);
    --tt-color: var(--light-grey);
    border: solid 1px var(--container-border-color);
  }

  &::after {
    position: relative;
    transition: opacity 1s;
    opacity: 0;
    content: "";
    border-width: 5px;
    border-style: solid;
    position: absolute;
    z-index: 1;
    pointer-events: none;

    top: -10px;
    left: 50%;
    border-color: transparent transparent var(--tt-bg-color) transparent;
  }

  &.hover,
  &.open {
    opacity: 1;
    pointer-events: unset;

    &:not(.dark)::after {
      opacity: 1;
    }
  }

  &.top {
    transform: translate(-50%, -100%);

    &::after {
      top: 100%;
      left: 50%;
      border-color: var(--tt-bg-color) transparent transparent transparent;
    }

  }

  &.right {
    transform: translate(0, -50%);

    &::after {
      top: calc(50% - 5px);
      left: -10px;
      border-color: transparent var(--tt-bg-color) transparent transparent;
    }
  }

  &.left {
    transform: translate(-100%, -50%);

    &::after {
      top: calc(50% - 5px);
      left: 100%;
      border-color: transparent transparent transparent var(--tt-bg-color);
    }
  }

  b {
    font-weight: var(--inter-medium);
  }
}
</style>
