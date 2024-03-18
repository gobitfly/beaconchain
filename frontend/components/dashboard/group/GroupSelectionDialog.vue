<script setup lang="ts">
import type { DynamicDialogInstance } from 'primevue/dynamicdialogoptions'
import { ref, onMounted, inject } from 'vue'

interface Props {
  groupId?: number
  selectedValidators?: number
  totalValidators?: number
}
const data = ref<Props>({})
const { t: $t } = useI18n()

const dialogRef = inject<Ref<DynamicDialogInstance>>('dialogRef')

onMounted(() => {
  if (dialogRef?.value?.options) {
    if (!dialogRef.value.options.props) {
      dialogRef.value.options.props = {}
    }
    dialogRef.value.options.props.dismissableMask = true
    dialogRef.value.options.props.modal = true
    dialogRef.value.options.props.header = $t('dashboard.group.selection.dialog.title')
  }
  data.value = dialogRef?.value.data
})

const closeDialog = (groupId?: number) => {
  dialogRef?.value.close(groupId)
}
</script>

<template>
  <div class="content">
    <div class="form">
      {{ $t('dashboard.group.selection.dialog.assign-group') }}
      <DashboardGroupSelection v-model="data.groupId" class="group-selection" />
    </div>
    <div class="footer">
      <b v-if="data.totalValidators"> {{ $t('dashboard.group.selection.dialog.validators-selected', {
        total: data.totalValidators
      }, data.selectedValidators ?? 0 ) }}</b>
      <Button
        :disabled="data.groupId === undefined"
        type="button"
        :label="$t('navigation.save')"
        @click="closeDialog(data.groupId)"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.content {
  display: flex;
  flex-direction: column;
  width: 350px;

  .form {
    flex-grow: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--padding);
    margin: 25px 0;

    .group-selection {
      width: 195px;
    }
  }

  .footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--padding);
  }
}
</style>
