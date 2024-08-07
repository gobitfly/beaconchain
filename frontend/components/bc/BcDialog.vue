<script lang="ts" setup>
import type Dialog from 'primevue/dialog'

const { width } = useWindowSize()
const { setTouchableElement } = useSwipe()

interface Props {
  header?: string
}
const props = defineProps<Props>()

const dialog = ref<{ container: HTMLElement } | undefined>()

const visible = defineModel<boolean>() // requires two way binding as both the parent (only the parent can open the modal) and the component itself (clicking outside the modal closes it) need to update the visibility

const position = computed(() => width.value <= 430 ? 'bottom' : 'center')

const onShow = () => {
  if (dialog.value?.container) {
    setTouchableElement(dialog.value?.container, () => {
      visible.value = false
      return true
    })
  }
}
</script>

<template>
  <Dialog
    ref="dialog"
    v-model:visible="visible"
    modal
    :header="props.header"
    :dismissable-mask="true"
    :draggable="false"
    :position="position"
    class="modal_container"
    :class="{ 'p-dialog-header-hidden': !props.header && !$slots.header }"
    @show="onShow"
  >
    <template #header>
      <slot name="header" />
    </template>
    <slot />
    <template #footer>
      <slot name="footer" />
    </template>
  </Dialog>
</template>
