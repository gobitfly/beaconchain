<script lang="ts" setup>
const { width } = useWindowSize()

interface Props {
  header?: string,
}
const props = defineProps<Props>()

const visible = defineModel<boolean>() // requires two way binding as both the parent (only the parent can open the modal) and the component itself (clicking outside the modal closes it) need to update the visibility

const isMobile = computed(() => width.value <= 430)
const position = computed(() => isMobile.value ? 'bottom' : 'center')
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    :header="props.header"
    :dismissable-mask="true"
    :closable="false"
    :draggable="false"
    :class="{
      'modal_container': true,
      'mobile_modal_container': isMobile,
      'p-dialog-header-hidden':!props.header && !$slots.header}"
    :position="position"
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

<style lang="scss" scoped>
  :global(.modal_container) {
    min-width: 375px;
  }

  :global(.mobile_modal_container) {
    margin-bottom: 0px;
  }
</style>
