<script setup lang="ts">

interface Props {
    text?: number,
    layout?: 'dark' | 'default'
    position?: 'top' | 'left' | 'right' | 'bottom'
}

const props = defineProps<Props>()
const bcTooltip = ref<HTMLElement | null >(null)
const isOpen = ref(false)

const classList = computed(() => {
  return [props.layout || 'default', props.position || 'bottom', isOpen.value ? 'open' : 'closed']
})

const handleClick = () => {
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
  <div ref="bcTooltip" class="bc-tooltip-wrapper" :class="classList" @click="handleClick()" @blur="isOpen = false">
    <slot />
    <div class="bc-tooltip" @click="$event.stopImmediatePropagation()">
      <slot name="tooltip">
        {{ text }}
      </slot>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.bc-tooltip-wrapper {
    position: relative;

    --tt-bg-color: var(--light-grey);
    --tt-color: var(--light-black);

    &.dark {
        --tt-bg-color: var(--light-black);
        --tt-color: var(--light-grey);
        .bc-tooltip{
            border: solid 1px var(--container-border-color);
        }
    }

    &::after {
        transition: opacity 1s;
        opacity: 0;
        content: "";
        border-width: 5px;
        border-style: solid;
        position: absolute;
        z-index: 1;
        pointer-events: none;

        inset-block-end: -20%;
        inset-inline-start: 40%;
        border-color: transparent transparent var(--tt-bg-color) transparent;
    }

    .bc-tooltip {
        opacity: 0;
        position: absolute;
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
        z-index: 1;

        inset-block-start: 120%;
        inset-inline-start: 50%;
        margin-inline-start: -60px;
    }

    &:hover,
    &:focus,
    &.open {

        &::after,
        .bc-tooltip {
            opacity: 1;
        }
    }
    &.open{
        &::after,
        .bc-tooltip {
            pointer-events: unset;
        }
    }

    &.top {
        &::after {
            inset-block-start: -20%;
            inset-inline-start: 40%;
            border-color: var(--tt-bg-color) transparent transparent transparent;
        }

        .bc-tooltip {
            inset-block-end: 120%;
            inset-inline-start: 50%;
            margin-inline-start: -60px;
        }
    }

    &.right {
        &::after {
            inset-block-start: 25%;
            inset-inline-end: -20%;
            border-color: transparent var(--tt-bg-color) transparent transparent;
        }

        .bc-tooltip {
            inset-block-end: 0%;
            inset-inline-start: 120%;
            min-height: 100%;
        }
    }

    &.left {
        &::after {
            inset-block-start: 25%;
            inset-inline-start: -20%;
            border-color: transparent transparent transparent var(--tt-bg-color);
        }

        .bc-tooltip {
            inset-block-end: 0%;
            inset-inline-end: 120%;
            min-height: 100%;
        }
    }
}
</style>
