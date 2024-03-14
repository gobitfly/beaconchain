<script setup lang="ts">
import type { DynamicDialogInstance } from 'primevue/dynamicdialogoptions'
import { ref, onMounted, inject } from 'vue'

const question = ref<string>('')
const dialogRef = inject<Ref<DynamicDialogInstance>>('dialogRef')

onMounted(() => {
  if (dialogRef?.value?.options) {
    if (!dialogRef.value.options.props) {
      dialogRef.value.options.props = {}
    }
    dialogRef.value.options.props.dismissableMask = true
    dialogRef.value.options.props.modal = true
  }
  question.value = dialogRef?.value.data.question
})

const closeDialog = (response: boolean) => {
  dialogRef?.value.close(response)
}
</script>

<template>
  <div class="content">
    <div class="question">
      {{ question }}
    </div>
    <div class="footer">
      <Button type="button" :label="$t('navigation.no')" @click="closeDialog(false)" />
      <Button type="button" :label="$t('navigation.yes')" @click="closeDialog(true)" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.content {
  display: flex;
  flex-direction: column;

  .question {
    flex-grow: 1;
    margin: var(--padding) 0;
  }

  .footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--padding);
  }
}
</style>
