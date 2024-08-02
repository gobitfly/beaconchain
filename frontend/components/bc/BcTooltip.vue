<script setup lang="ts">
import { useTooltipStore } from '~/stores/useTooltipStore'

interface Props {
  text?: string,
  title?: string,
  layout?: 'special' | 'default'
  position?: 'top' | 'left' | 'right' | 'bottom',
  hide?: boolean,
  tooltipClass?: string,
  fitContent?: boolean,
  renderTextAsHtml?: boolean,
  scrollContainer?: string // query selector for scrollable parent container
  dontOpenPermanently?: boolean
  hoverDelay?: number
}

const props = defineProps<Props>()
const bcTooltipOwner = ref<HTMLElement | null>(null)
const bcTooltip = ref<HTMLElement | null>(null)
let scrollParents: HTMLElement[] = []
const tooltipAddedTimeout = ref<NodeJS.Timeout | null>(null)
const { selected, doSelect } = useTooltipStore()
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
  return [props.layout || 'default', props.position || 'bottom', isOpen.value ? 'open' : 'closed', props.fitContent ? 'fit-content' : '']
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
  if (bcTooltip.value) {
    let centerX = -5 + Math.abs(left - rect.left) + rect.width / 2
    if (rect.width > ttWidth) {
      centerX = -5 + ttWidth / 2
    }
    let centerY = -5 + Math.abs(top - rect.top) + rect.height / 2
    if (rect.height > ttHeight) {
      centerY = -5 + ttHeight / 2
    }
    centerX = Math.max(5, Math.min(centerX, ttWidth - 5))
    centerY = Math.max(5, Math.min(centerY, ttHeight - 5))
    let afterLeft = centerX
    let afterTop = -10
    switch (props.position) {
      case 'bottom':
        break
      case 'left':
        afterLeft = ttWidth
        afterTop = centerY
        break
      case 'top':
        afterTop = ttHeight
        break
      case 'right':
        afterLeft = -10
        afterTop = centerY
        break
    }
    bcTooltip.value.style.setProperty('--tt-after-left', `${afterLeft}px`)
    bcTooltip.value.style.setProperty('--tt-after-top', `${afterTop}px`)
  }
}

const handleClick = () => {
  if (isSelected.value) {
    doSelect(null)
  } else if (canBeOpened.value) {
    if (props.dontOpenPermanently) {
      instantHover(true)
    } else {
      doSelect(bcTooltipOwner.value)
    }
    setPosition()
  }
}

const onHover = () => {
  if (canBeOpened.value && !selected.value) {
    if (props.hoverDelay) {
      bounceHover(true, false, false, props.hoverDelay)
    } else {
      instantHover(true)
      setPosition()
    }
  }
}

const doHide = (event?: Event) => {
  if (event?.target === bcTooltipOwner.value || isParent(bcTooltipOwner.value, event?.target as HTMLElement)) {
    return
  }
  removeParentListeners()
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

const addScrollParent = () => {
  removeScrollParent()
  scrollParents = findAllScrollParents(bcTooltipOwner.value)
  scrollParents.forEach(elem => elem.addEventListener('scroll', doHide))
}
const removeScrollParent = () => {
  scrollParents.forEach(elem => elem.removeEventListener('scroll', doHide))
}

watch(() => [props.title, props.text], () => {
  if (isOpen.value) {
    requestAnimationFrame(() => {
      setPosition()
    })
  }
})

const onWindowResize = () => {
  doHide()
}

watch(isOpen, (value) => {
  if (value) {
    setPosition()
    document.addEventListener('click', doHide)
    document.addEventListener('scroll', doHide)
    window.addEventListener('resize', onWindowResize)
    checkScrollListener(true)
    addScrollParent()
  }
})

function removeParentListeners () {
  document.removeEventListener('click', doHide)
  document.removeEventListener('scroll', doHide)
  window.removeEventListener('resize', onWindowResize)
  checkScrollListener(false)
  removeScrollParent()
}

onUnmounted(() => {
  removeParentListeners()
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
      <div class="bc-tooltip-wrapper" :style="pos" :class="tooltipClass">
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
              <template v-if="renderTextAsHtml && text">
                <!-- eslint-disable-next-line vue/no-v-html -->
                <span v-html="text" />
              </template>
              <template v-else>
                {{ text }}
              </template>
            </span>
          </slot>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

@keyframes fadeIn {
  0% {
    opacity: 0;
  }

  50% {
    opacity: 0;
  }

  100% {
    opacity: 1;
  }
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
  --tt-bg-color: var(--tooltip-background);
  --tt-color: var(--tooltip-text-color);
  --tt-after-left: unset;
  --tt-after-top: unset;
  position: relative;
  display: inline-flex;
  flex-wrap: wrap;
  opacity: 0;
  transition: opacity 1s;
  text-align: center;
  padding: 9px 12px;
  border-radius: var(--border-radius);
  background: var(--tt-bg-color);
  color: var(--tt-color);
  @include fonts.tooltip_text;
  pointer-events: none;
  max-width: 300px;

  &.special {
    --tt-bg-color: var(--light-grey-5);
    --tt-color: var(--light-black);
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

    top: var(--tt-after-top);
    left: var(--tt-after-left);
    border-color: transparent transparent var(--tt-bg-color) transparent;
  }

  &.hover,
  &.open {
    opacity: 1;
    pointer-events: unset;

    &:not(.special)::after {
      opacity: 1;
    }
  }

  &.top {
    &::after {
      border-color: var(--tt-bg-color) transparent transparent transparent;
    }

  }

  &.right {
    &::after {
      border-color: transparent var(--tt-bg-color) transparent transparent;
    }
  }

  &.left {
    &::after {
      border-color: transparent transparent transparent var(--tt-bg-color);
    }
  }

  :deep(.bold),
  :deep(b) {
    font-weight: var(--tooltip_text_bold_font_weight);
  }

  &:has(b):not(.fit-content) {
    min-width: 200px;
    text-align: left;
  }

  &.fit-content {
    min-width: max-content;
  }
}

.dark-mode{
  .bc-tooltip {
    &.special{
    --tt-bg-color: var(--light-black);
    --tt-color: var(--light-grey);
    }
  }
}
</style>
