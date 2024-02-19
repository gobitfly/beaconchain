<script lang="ts" setup>
interface Props {
  header?: string,
}
const props = defineProps<Props>()

const visible = defineModel<boolean>() // requires two way binding as both the parent (only the parent can open the modal) and the component itself (clicking outside the modal closes it) need to update the visibility
</script>

<template>
  <Dialog
    v-model:visible="visible"
    modal
    :header="props.header"
    :dismissable-mask="true"
    :closable="false"
    :draggable="false"
  >
    <slot />
  </Dialog>
</template>
