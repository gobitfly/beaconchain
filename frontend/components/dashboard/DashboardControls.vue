<script lang="ts" setup>
import {
  faDesktop,
  faEdit,
  faGear,
  faPeopleGroup,
  faShare,
  faTrash,
  faUsers,
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import {
  BcDialogConfirm,
  DashboardRenameModal,
  DashboardShareCodeModal,
  DashboardShareModal,
  RocketpoolToggle,
} from '#components'
import type {
  Dashboard, DashboardKey,
} from '~/types/dashboard'
import type {
  MenuBarButton, MenuBarEntry,
} from '~/types/menuBar'
import { API_PATH } from '~/types/customFetch'

interface Props {
  dashboardTitle?: string,
}
const props = defineProps<Props>()

const route = useRoute()
const isValidatorDashboard = route.name === 'dashboard-id'
const { isLoggedIn } = useUserStore()
const {
  dashboardKey,
  dashboardType,
  isPrivate,
  isPublic,
  isShared,
  publicEntities,
  setDashboardKey,
} = useDashboardKey()
const { refreshOverview } = useValidatorDashboardOverviewStore()
const {
  dashboards, getDashboardLabel, refreshDashboards, updateHash,
}
  = useUserDashboardStore()

const { t: $t } = useTranslation()
const { width } = useWindowSize()
const dialog = useDialog()
const { fetch } = useCustomFetch()

const isMobile = computed(() => width.value < 520)
const manageGroupsModalVisisble = ref(false)
const manageValidatorsModalVisisble = ref(false)

const manageButtons = computed<MenuBarEntry[] | undefined>(() => {
  if (isShared.value) {
    return undefined
  }

  const buttons: MenuBarEntry[] = []

  buttons.push({
    command: () => {
      manageGroupsModalVisisble.value = true
    },
    dropdown: false,
    faIcon: isMobile.value ? faPeopleGroup : undefined,
    label: $t('dashboard.validator.manage_groups'),
  })

  if (dashboardType.value === 'validator') {
    buttons.push({
      command: () => {
        manageValidatorsModalVisisble.value = true
      },
      dropdown: false,
      faIcon: isMobile.value ? faDesktop : undefined,
      highlight: !isMobile.value,
      label: $t('dashboard.validator.manage_validators'),
    })
  }

  if (isMobile.value && buttons.length > 1) {
    return [ {
      dropdown: true,
      highlight: true,
      items: buttons,
      label: $t('dashboard.header.manage'),
    } ]
  }

  return buttons
})

const shareDashboard = computed(() => {
  return dashboards.value?.validator_dashboards?.find((d) => {
    return (
      d.id === parseInt(dashboardKey.value)
      || d.public_ids?.find(p => p.public_id === dashboardKey.value)
    )
  })
})

const shareButtonOptions = computed(() => {
  const edit = isPrivate.value && !shareDashboard.value?.public_ids?.length

  const label = isMobile.value
    ? ''
    : !edit
        ? $t('dashboard.shared')
        : $t('dashboard.share')
  const icon = !edit ? faUsers : faShare
  const disabled = isShared.value || !dashboardKey.value
  return {
    disabled,
    edit,
    icon,
    label,
  }
})

const editButtons = computed<MenuBarEntry[]>(() => {
  const buttons: MenuBarButton[] = []

  buttons.push({ component: RocketpoolToggle })

  if (isPrivate.value) {
    buttons.push({
      command: editDashboard,
      faIcon: faEdit,
      label: $t('dashboard.rename_dashboard'),
    })
  }

  if (!shareButtonOptions.value.disabled) {
    buttons.push({
      command: share,
      faIcon: shareButtonOptions.value.icon,
      label: shareButtonOptions.value.edit
        ? $t('dashboard.share_dashboard')
        : $t('dashboard.shared_dashboard'),
    })
  }

  if (!isShared.value && dashboardKey.value) {
    buttons.push({
      command: onDelete,
      faIcon: faTrash,
      label: $t('dashboard.delete_dashboard'),
    })
  }

  return [ {
    dropdown: true,
    faIcon: faGear,
    items: buttons,
  } ]
})

const shareView = () => {
  const dashboardId = shareDashboard.value?.id
  dialog.open(DashboardShareCodeModal, {
    data: {
      dashboard: shareDashboard.value,
      dashboardKey: dashboardKey.value,
    },
    onClose: (options?: DynamicDialogCloseOptions) => {
      if (options?.data === 'DELETE') {
        if (isShared.value && dashboardId) {
          setDashboardKey(`${dashboardId}`)
        }
      }
      else if (options?.data) {
        shareEdit()
      }
    },
  })
}

const shareEdit = () => {
  dialog.open(DashboardShareModal, {
    data: { dashboard: shareDashboard.value },
    onClose: (options?: DynamicDialogCloseOptions) => {
      options?.data && shareView()
    },
  })
}

const share = () => {
  if (shareButtonOptions.value.edit) {
    shareEdit()
  }
  else {
    shareView()
  }
}

const deleteButtonOptions = computed(() => {
  const visible = !isShared.value

  const disabled = isPublic.value && publicEntities.value?.length === 0

  // private dashboards always get deleted, public dashboards only get cleared
  const deleteDashboard = isPrivate.value

  // we can only forward if there is something to forward to after a potential deletion
  const privateDashboardsCount = isLoggedIn.value
    ? (dashboards.value?.validator_dashboards?.length ?? 0)
    + (dashboards.value?.account_dashboards?.length ?? 0)
    : 0
  const forward = deleteDashboard
    ? privateDashboardsCount > 1
    : privateDashboardsCount > 0

  return {
    deleteDashboard,
    disabled,
    forward,
    visible,
  }
})

const onDelete = () => {
  const isDelete = deleteButtonOptions.value.deleteDashboard
  const dialogData = {
    noLabel: isDelete ? $t('dashboard.deletion.delete.no_label') : undefined,
    question: $t(
      isDelete
        ? 'dashboard.deletion.delete.text'
        : 'dashboard.deletion.clear.text',
      { dashboard: getDashboardLabel(dashboardKey.value, dashboardType.value) },
    ),
    severity: isDelete ? 'danger' : undefined,
    title: $t(
      isDelete
        ? 'dashboard.deletion.delete.title'
        : 'dashboard.deletion.clear.title',
    ),
    yesLabel: isDelete ? $t('dashboard.deletion.delete.yes_label') : undefined,
  }

  dialog.open(BcDialogConfirm, {
    data: dialogData,
    onClose: response =>
      response?.data
      && deleteAction(
        dashboardKey.value,
        deleteButtonOptions.value.deleteDashboard,
        deleteButtonOptions.value.forward,
      ),
  })
}

const deleteAction = async (
  key: DashboardKey,
  deleteDashboard: boolean,
  forward: boolean,
) => {
  if (deleteDashboard) {
    if (dashboardType.value === 'validator') {
      await fetch(
        API_PATH.DASHBOARD_DELETE_VALIDATOR,
        { body: { key } },
        { dashboardKey: key },
      )
    }
    else {
      await fetch(
        API_PATH.DASHBOARD_DELETE_ACCOUNT,
        { body: { key } },
        { dashboardKey: key },
      )
    }

    await refreshDashboards()
  }
  else if (!isLoggedIn.value) {
    // simply clear the public dashboard by emptying the hash
    updateHash(dashboardType.value, '')
    setDashboardKey('')
    return
  }

  if (forward) {
    // try to forward the user to a private dashboard
    let preferedDashboards: Dashboard[]
      = dashboards.value?.validator_dashboards ?? []
    let fallbackDashboards: Dashboard[]
      = dashboards.value?.account_dashboards ?? []
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

const title = computed(() => {
  return (
    props?.dashboardTitle
    || getDashboardLabel(
      dashboardKey.value,
      isValidatorDashboard ? 'validator' : 'account',
    )
  )
})

const editDashboard = () => {
  const list = isValidatorDashboard
    ? dashboards.value?.validator_dashboards
    : dashboards.value?.account_dashboards
  const dashboard = list?.find(d => `${d.id}` === dashboardKey.value)
  if (!dashboard) {
    return
  }
  dialog.open(DashboardRenameModal, {
    data: {
      dashboard,
      dashboardType: dashboardType.value,
    },
    onClose: (value?: DynamicDialogCloseOptions | undefined) => {
      if (value?.data === true) {
        refreshDashboards()
        refreshOverview(dashboardKey.value)
      }
    },
  })
}
</script>

<template>
  <DashboardGroupManagementModal v-model="manageGroupsModalVisisble" />
  <DashboardValidatorManagementModal
    v-if="dashboardType == 'validator'"
    v-model="manageValidatorsModalVisisble"
  />
  <div class="header-row">
    <div class="h1 dashboard-title">
      {{ title }}
    </div>
    <div class="action-button-container">
      <Button
        data-secondary
        class="share-button"
        :class="{ 'p-button-icon-only': !shareButtonOptions.label }"
        :disabled="shareButtonOptions.disabled"
        @click="share()"
      >
        {{ shareButtonOptions.label }}
        <FontAwesomeIcon :icon="shareButtonOptions.icon" />
      </Button>
      <BcMenuBar
        :buttons="editButtons"
        :align-right="isMobile"
      />
    </div>
    <BcMenuBar
      :buttons="manageButtons"
      :align-right="true"
    />
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";
@use "~/assets/css/fonts.scss";

.header-row {
  height: 30px;
  display: flex;
  gap: var(--padding);
  margin-bottom: var(--padding-large);
  @media (max-width: 519px) {
    gap: var(--padding-small);
  }

  .dashboard-title {
    @include utils.truncate-text;
  }

  .action-button-container {
    flex-grow: 1;
    display: flex;
    justify-content: flex-start;
    gap: var(--padding);
    @media (max-width: 519px) {
      justify-content: flex-end;
      gap: var(--padding-small);
    }

    .share-button {
      display: flex;
      gap: var(--padding-small);
    }
  }
}
</style>
