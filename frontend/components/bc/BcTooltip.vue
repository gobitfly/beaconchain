<script setup lang="ts">

interface Props {
  text?: string,
  layout?: 'dark' | 'default'
  position?: 'top' | 'left' | 'right' | 'bottom'
}

const props = defineProps<Props>()
const bcTooltip = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const pos = ref<{ top: string, left: string }>({ top: '0', left: '0' })

const classList = computed(() => {
  return [props.layout || 'default', props.position || 'bottom', isOpen.value ? 'open' : 'closed']
})

const handleClick = () => {
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

  isOpen.value = !isOpen.value
  if (!isOpen.value) {
    bcTooltip.value?.blur()
  }
}

const hide = (event: MouseEvent) => {
  if (event.target === bcTooltip.value || isParent(bcTooltip.value, event.target as HTMLElement)) {
    return
  }
  isOpen.value = false
  if (!isOpen.value) {
    bcTooltip.value?.blur()
  }
}

onMounted(() => {
  document.addEventListener('click', hide)
})

onUnmounted(() => {
  document.removeEventListener('click', hide)
})

</script>
<template>
  <div ref="bcTooltip" @click="handleClick()" @blur="isOpen = false">
    <slot />
    <Teleport v-if="isOpen" to="body">
      <div class="bc-tooltip-wrapper" :style="pos">
        <div class="bc-tooltip" :class="classList" @click="$event.stopImmediatePropagation()">
          <slot name="tooltip">
            {{ text }}
          </slot>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style lang="scss" scoped>
.bc-tooltip-wrapper {
  position: fixed;
  width: 1px;
  height: 1px;
  overflow: visible;
  z-index: 99999;

}

.bc-tooltip {

  --tt-bg-color: var(--light-grey);
  --tt-color: var(--light-black);

  position: relative;
  display: inline-flex;
  opacity: 0;
  transition: opacity 1s;
  text-align: center;
  padding: 9px 12px;
  min-width: 120px;
  border-radius: var(--border-radius);
  color: var(--tt-color);
  background: var(--tt-bg-color);
  font-family: var(--roboto-family);
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

  &:hover,
  &:focus,
  &.open {
    opacity: 1;

    &:not(.dark)::after {
      opacity: 1;
    }
  }

  &.open {
    pointer-events: unset;

    &::after {
      pointer-events: unset;
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
}
</style>
