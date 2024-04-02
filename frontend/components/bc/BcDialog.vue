<script lang="ts" setup>
const { width } = useWindowSize()

interface Props {
  header?: string,
}
const props = defineProps<Props>()

const visible = defineModel<boolean>() // requires two way binding as both the parent (only the parent can open the modal) and the component itself (clicking outside the modal closes it) need to update the visibility

const position = computed(() => width.value <= 430 ? 'bottom' : 'center')
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    :header="props.header"
    :dismissable-mask="true"
    :draggable="false"
    :position="position"
    class="modal_container"
    :class="{'p-dialog-header-hidden':!props.header && !$slots.header}"
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
