<script lang="ts" setup>
interface Props {
  header?: string,
}
const props = defineProps<Props>()

const visible = defineModel<boolean>()
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
