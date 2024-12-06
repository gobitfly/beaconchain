<script lang="ts" setup>
import type { NotificationSettingsDashboardsTableRow } from '~/types/api/notifications'

const { t: $t } = useTranslation()
const {
  close,
  props,
} = useBcDialog<NotificationSettingsDashboardsTableRow>()

const idCancel = useId()
// necessary as focus management with `primevues dialog` is not working properly
const lastFocusedElement = ref()
onBeforeMount(() => {
  lastFocusedElement.value = document.activeElement
})
onMounted(() => {
  document.getElementById(idCancel)?.focus()
})
onUnmounted(() => {
  lastFocusedElement.value?.focus()
})
const emit = defineEmits<{
  (e: 'delete',
    payload: Pick<
      NotificationSettingsDashboardsTableRow,
      | 'dashboard_id'
      | 'group_id'
      | 'is_account_dashboard'
      | 'settings'
    >
  ): void,
}>()
const handleDelete = () => {
  if (
    typeof props.value?.dashboard_id === 'number'
    && typeof props.value?.group_id === 'number'
  ) {
    emit('delete', {
      dashboard_id: props.value.dashboard_id,
      group_id: props.value.group_id,
      is_account_dashboard: props.value.is_account_dashboard,
      settings: props.value.settings,
    })
  }
  close()
}
</script>

<template>
  <div
    v-focustrap
    @keydown.esc.stop="close"
  >
    <BcText tag="h2" variant="lg">
      {{ $t('notifications.dashboards.dialog.delete_all_notifications.heading') }}
    </BcText>
    <p class="notifications-management-modal-dashboards-delete__content">
      {{ $t('notifications.dashboards.dialog.delete_all_notifications.paragraph') }}
    </p>
    <div class="notifications-management-modal-dashboards-delete__footer">
      <BcButton
        :id="idCancel"
        variant="secondary"
        @click="close"
      >
        {{ $t("navigation.cancel") }}
      </BcButton>
      <BcButton
        class="notifications-management-dialog-webhook__primary-button"
        @click="handleDelete"
      >
        {{ $t("navigation.delete") }}
      </BcButton>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.notifications-management-modal-dashboards-delete__footer {
  display: flex;
  justify-content: flex-end;
  margin-top: var(--padding-small);
}
</style>
