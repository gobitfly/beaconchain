<script lang="ts" setup>
import {
  faShare,
  faUsers,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { BcDialogConfirm } from '#components'
import type { DashboardKey } from '~/types/dashboard'
import type { MenuBarEntry } from '~/types/menuBar'

const { dashboardKey, isPublic, setDashboardKey, dashboardType, publicEntities, removeEntities } = useDashboardKey()
const { refreshDashboards, dashboards, getDashboardLabel } = useUserDashboardStore()

const { t: $t } = useI18n()
const { width } = useWindowSize()
const dialog = useDialog()
const { fetch } = useCustomFetch()

const manageGroupsModalVisisble = ref(false)
const manageValidatorsModalVisisble = ref(false)

const manageButtons = computed<MenuBarEntry[] | undefined>(() => {
  const buttons: MenuBarEntry[] = []

  buttons.push({
    dropdown: false,
    label: $t('dashboard.validator.manage_groups'),
    command: () => { manageGroupsModalVisisble.value = true }
  })

  if (dashboardType.value === 'validator') {
    buttons.push(
      {
        dropdown: false,
        label: $t('dashboard.validator.manage_validators'),
        command: () => { manageValidatorsModalVisisble.value = true }
      }
    )
  }

  if (width.value < 520 && buttons.length > 1) {
    return [
      {
        label: 'Manage',
        dropdown: true,
        items: buttons
      }
    ]
  }

  return buttons
})

const shareButtonOptions = computed(() => {
  const label = isPublic.value ? $t('dashboard.shared') : $t('dashboard.share')
  const icon = isPublic.value ? faUsers : faShare
  return { label, icon }
})

const share = () => {
  alert('Not implemented yet')
}

const deleteButtonOptions = computed(() => {
  const disabled = isPublic.value && publicEntities.value?.length === 0

  return { disabled }
})

const onDelete = () => {
  dialog.open(BcDialogConfirm, {
    props: {
      header: $t('dashboard.deletion.title')
    },
    onClose: response => response?.data && deleteDashboard(dashboardKey.value),
    data: {
      question: $t('dashboard.deletion.text', { dashboard: getDashboardLabel(dashboardKey.value, dashboardType.value) })
    }
  })
}

const deleteDashboard = async (key: DashboardKey) => {
  if (isPublic.value) {
    if (publicEntities.value?.length > 0) {
      removeEntities(publicEntities.value)
    }
    return
  }

  if (dashboardType.value === 'validator') {
    await fetch(API_PATH.DASHBOARD_DELETE_VALIDATOR, { body: { key } }, { dashboardKey: key })
  } else {
    await fetch(API_PATH.DASHBOARD_DELETE_ACCOUNT, { body: { key } }, { dashboardKey: key })
  }

  await refreshDashboards()

  let preferedDashboards = dashboards.value?.validator_dashboards ?? []
  let fallbackDashboards = dashboards.value?.account_dashboards ?? []
  let fallbackUrl = '/account-dashboard/'
  if (dashboardType.value === 'account') {
    preferedDashboards = dashboards.value?.account_dashboards ?? []
    fallbackDashboards = dashboards.value?.validator_dashboards ?? []
    fallbackUrl = '/dashboard/'
  }

  // forward user to another dashboard (if possible)
  if ((preferedDashboards?.length ?? 0) > 0) {
    setDashboardKey(`${preferedDashboards[0].id}`)
    return
  }

  if ((fallbackDashboards.length ?? 0) > 0) {
    await navigateTo(`${fallbackUrl}${fallbackDashboards[0].id}`)
    return
  }

  // no other dashboard available, forward to creation screen
  setDashboardKey('')
}
</script>

<template>
  <DashboardGroupManagementModal v-model="manageGroupsModalVisisble" />
  <DashboardValidatorManagementModal v-if="dashboardType=='validator'" v-model="manageValidatorsModalVisisble" />
  <div class="header-row">
    <div class="action-button-container">
      <Button class="share-button" :disabled="isPublic" @click="share()">
        {{ shareButtonOptions.label }}<FontAwesomeIcon :icon="shareButtonOptions.icon" />
      </Button>
      <Button class="p-button-icon-only" :disabled="deleteButtonOptions.disabled" @click="onDelete()">
        <FontAwesomeIcon :icon="faTrash" />
      </Button>
    </div>
    <Menubar v-if="manageButtons" :model="manageButtons" breakpoint="0px" class="right-aligned-submenu">
      <template #item="{ item }">
        <span class="button-content pointer">
          <span class="text">{{ item.label }}</span>
          <IconChevron v-if="item.dropdown" class="toggle" direction="bottom" />
        </span>
      </template>
    </Menubar>
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

.header-row {
  height: 30px;
  display: flex;
  justify-content: space-between;
  gap: var(--padding);
  margin-bottom: var(--padding);

  .action-button-container{
    display: flex;
    gap: var(--padding);

    .share-button{
      display: flex;
      gap: var(--padding-small);
    }
  }

  :deep(.p-menubar .p-menubar-root-list) {
    >.p-menuitem{
      height: 30px;
      color: var(--text-color-inverted);
      background: var(--button-color-active);
      font-weight: var(--standard_text_medium_font_weight);
      border-color: var(--button-color-active);

      >.p-menuitem-content {
        padding: 0;

        >.button-content{
          display: flex;
          align-items: center;
          gap: 7px;
          padding: 7px 17px;

          .pointer {
            cursor: pointer;
          }
        }
      }

      >.p-submenu-list {
        font-weight: var(--standard_text_font_weight);

        >.p-menuitem .button-content{
          gap: 0;
          padding: 0;

          .text {
            @include utils.truncate-text;
          }
        }
      }

      &:not(.p-highlight):not(.p-disabled) > .p-menuitem-content:hover {
        background: var(--button-color-hover);
      }
    }
  }
}
</style>
