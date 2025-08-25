<script lang="ts" setup>
import type { DynamicDialogCloseOptions } from 'primevue/dynamicdialogoptions'
import {
  LazyBcDialogConfirm,
  LazyDashboardRenameModal,
  LazyDashboardShareCodeModal,
  LazyDashboardShareModal,
} from '#components'
import type {
  MenuBarButton, MenuBarEntry,
} from '~/types/menuBar'
import type { Icon } from '~/components/bc/icon/BcIcon.vue'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'
import type { ValidatorDashboard } from '~/types/api/dashboard'

const props = defineProps<{
  dashboardTitle: string,
  validatorDashboards: null | ValidatorDashboard[],
}>()

const { isLoggedIn } = useUser()
// const {
//   dashboardKey,
//   dashboardType,
//   isGuestDashboard,
//   isPrivateDashboard,
//   isSharedDashboard,
//   publicEntities,
//   setDashboardKey,
// } = useDashboardKey()

// const { refreshOverview } = useValidatorDashboardOverviewStore()

const { t: $t } = useTranslation()
const { width } = useWindowSize()
const dialog = useDialog()
// const { fetch } = useCustomFetch()

const isMobile = computed(() => width.value < 520)
const manageGroupsModalVisisble = ref(false)

const isVisibleManagementModal = defineModel<boolean>('isVisibleManagementModal')
const {
  hasValidators,
  key,
  // validators,
  name,
  navigateToDashboard,
  // publicId,
  variant,
} = useDashboard()
const manageButtons = computed<MenuBarEntry[]>(() => {
  if (variant.value === 'shared-dashboard') {
    return []
  }

  const buttons: MenuBarEntry[] = []

  buttons.push({
    command: () => {
      manageGroupsModalVisisble.value = true
    },
    dropdown: false,
    faIcon: isMobile.value ? 'people-group' : undefined,
    label: $t('dashboard.validator.manage_groups'),
  })

  buttons.push({
    command: () => {
      isVisibleManagementModal.value = true
    },
    dropdown: false,
    faIcon: isMobile.value ? 'desktop' : undefined,
    highlight: !isMobile.value,
    label: $t('dashboard.validator.manage_validators'),
  })

  if (isMobile.value && buttons.length > 1) {
    return [ {
      dropdown: true,
      highlight: true,
      items: buttons.reverse(), // In the dropdown we want the button sorting to be reversed
      label: $t('dashboard.header.manage'),
    } ]
  }

  return buttons
})

const shareDashboard = computed(() => {
  return props.validatorDashboards?.find((d) => {
    return (
      d.id === parseInt(key.value ?? '')
      || d.public_ids?.find(p => p.public_id === key.value)
    )
  })
})

const shareButtonOptions = computed(() => {
  const edit = variant.value === 'private-dashboard' && !shareDashboard.value?.public_ids?.length

  const label = isMobile.value
    ? ''
    : !edit
        ? $t('dashboard.shared')
        : $t('dashboard.share')
  const icon: Icon = !edit ? 'people-group' : 'share'
  const disabled = variant.value === 'shared-dashboard' || !key.value
  return {
    disabled,
    edit,
    icon,
    label,
  }
})

const editButtons = computed<MenuBarEntry[]>(() => {
  const buttons: MenuBarButton[] = []

  if (variant.value === 'private-dashboard') {
    buttons.push({
      command: editDashboard,
      faIcon: 'edit',
      label: $t('dashboard.rename_dashboard'),
    })
  }

  if (!shareButtonOptions.value.disabled) {
    buttons.push({
      command: share,
      faIcon: shareButtonOptions.value.icon as Icon,
      label: shareButtonOptions.value.edit
        ? $t('dashboard.share_dashboard')
        : $t('dashboard.shared_dashboard'),
    })
  }

  if (variant.value !== 'shared-dashboard' && key.value) {
    buttons.push({
      command: onDelete,
      faIcon: 'trash',
      label: $t('dashboard.delete_dashboard'),
    })
  }

  return [ {
    dropdown: true,
    faIcon: 'gear',
    items: buttons,
  } ]
})

const shareView = () => {
  const dashboardId = shareDashboard.value?.id
  dialog.open(LazyDashboardShareCodeModal, {
    data: {
      dashboard: shareDashboard.value,
      dashboardKey: key.value,
    },
    onClose: (options?: DynamicDialogCloseOptions) => {
      if (options?.data === 'DELETE') {
        if (variant.value === 'shared-dashboard' && dashboardId) {
          navigateToDashboard(`${dashboardId}`)
        }
      }
      else if (options?.data) {
        shareEdit()
      }
    },
  })
}

const shareEdit = () => {
  dialog.open(LazyDashboardShareModal, {
    data: { dashboard: shareDashboard.value },
    onClose: (options?: DynamicDialogCloseOptions) => {
      if (options?.data) {
        shareView()
      }
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
  const visible = variant.value !== 'shared-dashboard'

  const disabled = variant.value === 'guest-dashboard' && hasValidators.value

  // private dashboards always get deleted, guest dashboards only get cleared
  const deleteDashboard = variant.value === 'private-dashboard'

  // we can only forward if there is something to forward to after a potential deletion
  const privateDashboardsCount = isLoggedIn.value
    ? (props.validatorDashboards?.length ?? 0)
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
      { dashboard: props.dashboardTitle },
    ),
    severity: isDelete ? 'danger' : undefined,
    title: $t(
      isDelete
        ? 'dashboard.deletion.delete.title'
        : 'dashboard.deletion.clear.title',
    ),
    yesLabel: isDelete ? $t('dashboard.deletion.delete.yes_label') : undefined,
  }

  dialog.open(LazyBcDialogConfirm, {
    data: dialogData,
    // onClose: response =>
    //   response?.data
    //   && deleteAction(
    //     key.value,
    //     deleteButtonOptions.value.deleteDashboard,
    //     deleteButtonOptions.value.forward,
    //   ),
  })
}

// const deleteAction = async (
//   key: DashboardKey,
//   deleteDashboard: boolean,
//   forward: boolean,
// ) => {

// TODO deletion of DBs

// if (deleteDashboard) {
//   await fetch(
//     'DASHBOARD_DELETE_VALIDATOR',
//     { body: { key } },
//     { dashboardKey: key },
//   )

//   await refreshDashboards()
// }
// else if (!isLoggedIn.value) {
//   // simply clear the guest dashboard by emptying the key
//   updateGuestDashboardKey(dashboardType.value, '')
//   setDashboardKey('')
//   return
// }

//   if (forward) {
//     // try to forward the user to a private dashboard
//     let preferedDashboards: Dashboard[]
//       = dashboards.value?.validator_dashboards ?? []
//     let fallbackDashboards: Dashboard[]
//       = dashboards.value?.account_dashboards ?? []
//     let fallbackUrl = '/account-dashboard/'
//     if (dashboardType.value === 'account') {
//       preferedDashboards = dashboards.value?.account_dashboards ?? []
//       fallbackDashboards = dashboards.value?.validator_dashboards ?? []
//       fallbackUrl = '/dashboard/'
//     }

//     if ((preferedDashboards?.length ?? 0) > 0) {
//       setDashboardKey(`${preferedDashboards[0].id}`)
//       return
//     }

//     if ((fallbackDashboards.length ?? 0) > 0) {
//       await navigateTo(`${fallbackUrl}${fallbackDashboards[0].id}`)
//       return
//     }
//   }

//   // no private dashboard available, forward to creation screen
//   setDashboardKey('')
// }

const editDashboard = () => {
  const list = props.validatorDashboards
  const dashboard = list?.find(d => `${d.id}` === key.value)
  if (!dashboard) {
    return
  }
  dialog.open(LazyDashboardRenameModal, {
    data: {
      dashboard,
    },
    onClose: (value?: DynamicDialogCloseOptions | undefined) => {
      if (value?.data === true) {
        // refreshDashboards()
      }
    },
  })
}
const emit = defineEmits<{
  (e: 'change-validators', value: string[]): void,
  (e: 'change-groups', value: VDBOverviewGroup[]): void,
}>()
</script>

<template>
  <DashboardGroupManagementModal
    v-model="manageGroupsModalVisisble"
    @change-groups="emit('change-groups', $event)"
  />
  <LazyDashboardValidatorManagementModal
    v-if="isVisibleManagementModal"
    v-model="isVisibleManagementModal"
    @change-validators="emit('change-validators', $event)"
  />
  <div class="header-row">
    <div class="h1 dashboard-title">
      {{ name }}
    </div>
    <div class="action-button-container">
      <Button
        severity="secondary"
        class="share-button"
        :class="{ 'p-button-icon-only': !shareButtonOptions.label }"
        :disabled="shareButtonOptions.disabled"
        @click="share()"
      >
        {{ shareButtonOptions.label }}
        <BcIcon :name="shareButtonOptions.icon" />
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
