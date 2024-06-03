<script lang="ts" setup>
import {
  faShare,
  faUsers,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import { BcDialogConfirm, DashboardShareModal, DashboardShareCodeModal } from '#components'
import type { DashboardKey } from '~/types/dashboard'
import type { MenuBarEntry } from '~/types/menuBar'
import { API_PATH } from '~/types/customFetch'

const { isLoggedIn } = useUserStore()
const { dashboardKey, isPublic, isPrivate, isShared, setDashboardKey, dashboardType, publicEntities } = useDashboardKey()
const { refreshDashboards, dashboards, getDashboardLabel, updateHash } = useUserDashboardStore()

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

const shareDashboard = computed(() => {
  return dashboards.value?.validator_dashboards?.find((d) => {
    return d.id === parseInt(dashboardKey.value) || d.public_ids?.find(p => p.public_id === dashboardKey.value)
  })
})

const shareButtonOptions = computed(() => {
  const edit = isPrivate.value && !shareDashboard.value?.public_ids?.length

  const label = !edit ? $t('dashboard.shared') : $t('dashboard.share')
  const icon = !edit ? faUsers : faShare
  return { label, icon, edit }
})

const shareView = () => {
  const dashboardId = shareDashboard.value?.id
  dialog.open(DashboardShareCodeModal, {
    data: { dashboard: shareDashboard.value, dashboardKey: dashboardKey.value },
    onClose: (options?: DynamicDialogCloseOptions) => {
      if (options?.data === 'DELETE') {
        if (isShared.value && dashboardId) {
          setDashboardKey(`${dashboardId}`)
        }
      } else if (options?.data) {
        shareEdit()
      }
    }
  })
}

const shareEdit = () => {
  dialog.open(DashboardShareModal, { data: { dashboard: shareDashboard.value }, onClose: (options?: DynamicDialogCloseOptions) => { options?.data && shareView() } })
}

const share = () => {
  if (shareButtonOptions.value.edit) {
    shareEdit()
  } else {
    shareView()
  }
}

const deleteButtonOptions = computed(() => {
  const disabled = isPublic.value && publicEntities.value?.length === 0

  // private dashboards always get deleted, public dashboards only get cleared
  const deleteDashboard = isPrivate.value

  // we can only forward if there is something to forward to after a potential deletion
  const privateDashboardsCount = isLoggedIn.value ? ((dashboards.value?.validator_dashboards?.length ?? 0) + (dashboards.value?.account_dashboards?.length ?? 0)) : 0
  const forward = deleteDashboard ? (privateDashboardsCount > 1) : (privateDashboardsCount > 0)

  return { disabled, deleteDashboard, forward }
})

const onDelete = () => {
  const languageKey = deleteButtonOptions.value.deleteDashboard ? 'dashboard.deletion.delete_text' : 'dashboard.deletion.clear_text'

  dialog.open(BcDialogConfirm, {
    props: {
      header: $t('dashboard.deletion.title')
    },
    onClose: response => response?.data && deleteAction(dashboardKey.value, deleteButtonOptions.value.deleteDashboard, deleteButtonOptions.value.forward),
    data: {
      question: $t(languageKey, { dashboard: getDashboardLabel(dashboardKey.value, dashboardType.value) })
    }
  })
}

const deleteAction = async (key: DashboardKey, deleteDashboard: boolean, forward: boolean) => {
  if (deleteDashboard) {
    if (dashboardType.value === 'validator') {
      await fetch(API_PATH.DASHBOARD_EDIT_VALIDATOR, { body: { key }, method: 'DELETE' }, { dashboardKey: key })
    } else {
      await fetch(API_PATH.DASHBOARD_EDIT_ACCOUNT, { body: { key }, method: 'DELETE' }, { dashboardKey: key })
    }

    await refreshDashboards()
  } else if (!isLoggedIn.value) {
    // simply clear the public dashboard by emptying the hash
    updateHash(dashboardType.value, '')
    setDashboardKey('')
    return
  }

  if (forward) {
    // try to forward the user to a private dashboard
    let preferedDashboards = dashboards.value?.validator_dashboards ?? []
    let fallbackDashboards = dashboards.value?.account_dashboards ?? []
    let fallbackUrl = '/account-dashboard/'
    if (dashboardType.value === 'account') {
      preferedDashboards = dashboards.value?.account_dashboards ?? []
      fallbackDashboards = dashboards.value?.validator_dashboards ?? []
      fallbackUrl = '/dashboard/'
    }

    if ((preferedDashboards?.length ?? 0) > 0) {
      setDashboardKey(`${preferedDashboards[0].id}`)
      return
    }

    if ((fallbackDashboards.length ?? 0) > 0) {
      await navigateTo(`${fallbackUrl}${fallbackDashboards[0].id}`)
      return
    }
  }

  // no private dashboard available, forward to creation screen
  setDashboardKey('')
}
</script>

<template>
  <DashboardGroupManagementModal v-model="manageGroupsModalVisisble" />
  <DashboardValidatorManagementModal v-if="dashboardType=='validator'" v-model="manageValidatorsModalVisisble" />
  <div class="header-row">
    <div class="action-button-container">
      <Button class="share-button" :disabled="!dashboardKey" @click="share()">
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
@use '~/assets/css/fonts.scss';

.header-row {
  height: 30px;
  display: flex;
  justify-content: space-between;
  gap: var(--padding);
  margin-bottom: var(--padding-large);

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
      color: var(--text-color-inverted);
      background: var(--button-color-active);
      border-color: var(--button-color-active);

      >.p-menuitem-content {
        margin-top: 1px;
        .button-content{
          .toggle {
            margin-left: var(--padding);
          }
        }
      }

      >.p-submenu-list {
        font-weight: var(--standard_text_font_weight);
      }

      &:not(.p-highlight):not(.p-disabled) > .p-menuitem-content:hover {
        background: var(--button-color-hover);
      }
    }
  }
}
</style>
