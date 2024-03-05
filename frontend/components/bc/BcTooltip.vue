<script setup lang="ts">
import { useTooltipStore } from '~/stores/useTooltipStore'

interface Props {
  text?: string,
  title?: string,
  layout?: 'dark' | 'default'
  position?: 'top' | 'left' | 'right' | 'bottom',
  hide?: boolean,
  scrollContainer?: string // query selector for scrollable parent container
}

const props = defineProps<Props>()
const bcTooltipOwner = ref<HTMLElement | null>(null)
const bcTooltip = ref<HTMLElement | null>(null)
const tooltipAddedTimeout = ref<NodeJS.Timeout | null>(null)
const ttStore = useTooltipStore()
const { doSelect } = ttStore
const { selected } = storeToRefs(ttStore)
const { width, height } = useWindowSize()

// this const will be avaiable on template
const slots = useSlots()

const hasContent = computed(() => !!slots.tooltip || !!props.text)
const canBeOpened = computed(() => !props.hide && hasContent.value)

const { value: hover, bounce: bounceHover, instant: instantHover } = useDebounceValue<boolean>(false, 50)
const { value: hoverTooltip, bounce: bounceHoverTooltip, instant: instantHoverTooltip } = useDebounceValue<boolean>(false, 50)
const isSelected = computed(() => !!bcTooltipOwner.value && selected.value === bcTooltipOwner.value)
const isOpen = computed(() => isSelected.value || hover.value || hoverTooltip.value)

const pos = ref<{ top: string, left: string }>({ top: '0', left: '0' })

const classList = computed(() => {
  return [props.layout || 'default', props.position || 'bottom', isOpen.value ? 'open' : 'closed']
})

const setPosition = () => {
  if (tooltipAddedTimeout.value) {
    clearTimeout(tooltipAddedTimeout.value)
    tooltipAddedTimeout.value = null
  }
  if (!isSelected.value && !hover.value) {
    return
  }
  const rect = bcTooltipOwner.value?.getBoundingClientRect()
  const tt = bcTooltip.value?.getBoundingClientRect?.()
  if (!rect) {
    return
  }
  if (!tt) {
    // we need to wait for the tt to be added to the dome to get it's measure, but we set the pos at an estimated value until then
    tooltipAddedTimeout.value = setTimeout(setPosition, 10)
  }
  const ttWidth = tt?.width ?? 100
  const ttHeight = tt?.height ?? 60
  const padding = 4
  let top = rect.bottom + padding
  let left = rect.left + rect.width / 2 - ttWidth / 2
  switch (props.position) {
    case 'left':
      left = rect.left - padding - ttWidth
      top = rect.top + rect.height / 2 - ttHeight / 2
      break
    case 'top':
      top = rect.top - padding - ttHeight
      break
    case 'right':
      left = rect.right + padding
      top = rect.top + rect.height / 2 - ttHeight / 2
      break
  }
  left = Math.max(0, Math.min(left, (width.value - ttWidth)))
  top = Math.max(0, Math.min(top, (height.value - ttHeight)))
  pos.value = { top: `${top}px`, left: `${left}px` }
}

const handleClick = () => {
  if (isSelected.value) {
    doSelect(null)
  } else if (canBeOpened.value) {
    doSelect(bcTooltipOwner.value)
    setPosition()
  }
}

const onHover = () => {
  if (canBeOpened.value && !selected.value) {
    instantHover(true)
    setPosition()
  }
}

const doHide = (event: Event) => {
  if (event.target === bcTooltipOwner.value || isParent(bcTooltipOwner.value, event.target as HTMLElement)) {
    return
  }
  if (isSelected.value) {
    doSelect(null)
  }
  instantHover(false)
  if (!isOpen.value) {
    bcTooltipOwner.value?.blur()
  }
}

const checkScrollListener = (add: boolean) => {
  if (props.scrollContainer) {
    const container = document.querySelector(props.scrollContainer)
    if (container) {
      if (add) {
        container.addEventListener('scroll', doHide)
      } else {
        container.removeEventListener('scroll', doHide)
      }
    }
  }
}

onMounted(() => {
  document.addEventListener('click', doHide)
  document.addEventListener('scroll', doHide)
  checkScrollListener(true)
})

onUnmounted(() => {
  document.removeEventListener('click', doHide)
  document.removeEventListener('scroll', doHide)
  checkScrollListener(false)
  if (isSelected.value) {
    doSelect(null)
  }
})

</script>
<template>
  <div
    ref="bcTooltipOwner"
    class="slot_container"
    @mouseover="onHover()"
    @mouseleave="bounceHover(false, false, true)"
    @click="handleClick()"
    @blur="bounceHover(false, false, true)"
  >
    <slot />
    <Teleport v-if="isOpen" to="body">
      <div class="bc-tooltip-wrapper" :style="pos">
        <div
          ref="bcTooltip"
          class="bc-tooltip"
          :class="classList"
          @click="$event.stopImmediatePropagation()"
          @mouseover="instantHoverTooltip(true)"
          @mouseleave="bounceHoverTooltip(false, false, true)"
        >
          <slot name="tooltip">
            <span>
              <b v-if="props.title">
                {{ props.title }}
              </b>
              {{ text }}
            </span>
          </slot>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style lang="scss" scoped>

@keyframes fadeIn {
  0% { opacity: 0; }
  50% { opacity: 0; }
  100% { opacity: 1; }
}
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
  opacity: 1;
  animation: fadeIn 100ms;
}

.bc-tooltip {
  position: relative;
  display: inline-flex;
  flex-wrap: wrap;
  opacity: 0;
  transition: opacity 1s;
  text-align: center;
  padding: 9px 12px;
  min-width: 120px;
  border-radius: var(--border-radius);
  color: var(--tooltip-text-color);
  background: var(--tooltip-background);
  font-family: var(--inter-family);
  font-weight: var(--inter-light);
  font-size: 10px;
  pointer-events: none;

  &.dark {
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
    left: calc(50% - 5px);
    border-color: transparent transparent var(--tooltip-background) transparent;
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
    &::after {
      top: 100%;
      left: calc(50% - 5px);
      border-color: var(--tooltip-background) transparent transparent transparent;
    }

  }

  &.right {
    &::after {
      top: calc(50% - 5px);
      left: -10px;
      border-color: transparent var(--tooltip-background) transparent transparent;
    }
  }

  &.left {
    &::after {
      top: calc(50% - 5px);
      left: 100%;
      border-color: transparent transparent transparent var(--tooltip-background);
    }
  }

  :deep(b) {
    font-weight: bold;
    font-weight: var(--inter-medium);
  }

  &:has(b) {
    min-width: 200px;
    text-align: left;
  }
}
</style>
