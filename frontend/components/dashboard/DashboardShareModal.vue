<script lang="ts" setup>
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { ValidatorDashboard } from '~/types/api/dashboard'
import { API_PATH } from '~/types/customFetch'

interface Props {
  dashboard: ValidatorDashboard; // Currently only validator dashboards are supported
}
const { props, dialogRef } = useBcDialog<Props>()
const { t: $t } = useI18n()
const { refreshDashboards } = useUserDashboardStore()
const { fetch } = useCustomFetch()

const dashboardName = ref('')
const shareGroups = ref(true)
const isUpdating = ref(false)
const isNew = ref(true)
const { user } = useUserStore()

const isPremiumUser = computed(() => !!user.value?.premium_perks?.share_custom_dashboards)

watch(props, (p) => {
  if (p) {
    // We currently only want to use one public id
    shareGroups.value = isPremiumUser.value && !!p.dashboard.public_ids?.[0]?.share_settings.group_names
    isNew.value = !p.dashboard.public_ids?.[0]
    if (isNew.value) {
      dashboardName.value = props.value?.dashboard?.name ?? ''
    } else {
      dashboardName.value = p.dashboard.public_ids?.[0]?.name ?? ''
    }
  }
}, { immediate: true })

const add = async () => {
  isUpdating.value = true
  await fetch(API_PATH.DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID, { body: { name: dashboardName.value, share_settings: { group_names: shareGroups.value } } }, { dashboardKey: `${props.value?.dashboard.id}` })
  await refreshDashboards()
  dialogRef?.value?.close(true)
  isUpdating.value = false
}

const edit = async () => {
  isUpdating.value = true
  const publicId = `${props.value?.dashboard.public_ids?.[0]?.public_id}`
  await fetch(API_PATH.DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID, { body: { name: dashboardName.value, share_settings: { group_names: shareGroups.value } } }, { dashboardKey: `${props.value?.dashboard.id}`, publicId })
  await refreshDashboards()
  dialogRef?.value?.close(true)
  isUpdating.value = false
}

const publishDisabled = computed(() => {
  return isUpdating.value || !REGEXP_VALID_NAME.test(dashboardName.value)
})

const share = () => {
  dashboardName.value = removeLeadingAndTrailingWhitespace(dashboardName.value)
  if (publishDisabled.value) {
    return
  }

  if (props.value?.dashboard.public_ids?.[0]?.public_id) {
    edit()
  } else {
    add()
  }
}

const shareGroupTooltip = computed(() => {
  return formatMultiPartSpan($t, 'dashboard.share_dialog.setting.group.tooltip', [undefined, 'bold', undefined])
})

</script>

<template>
  <div class="share-dashboard-modal-container">
    <div class="content">
      <label for="dashboardName" class="medium">{{ $t('dashboard.share_dialog.setting.name.label') }}</label>
      <InputText
        id="dashboardName"
        v-model="dashboardName"
        :placeholder="$t('dashboard.share_dialog.setting.name.placeholder')"
        class="input-field"
        @keypress.enter="share"
      />
      <div class="share-setting">
        <Checkbox id="shareGroup" v-model="shareGroups" :binary="true" :disabled="!isPremiumUser" />
        <label for="shareGroup" :class="{'text-disabled':!isPremiumUser}">{{ $t('dashboard.share_dialog.setting.group.label') }}</label>

        <BcTooltip
          position="top"
          tooltip-class="share-dialog-setting-tooltip"
          :text="shareGroupTooltip"
          :render-text-as-html="true"
        >
          <FontAwesomeIcon :icon="faInfoCircle" />
        </BcTooltip>
        <BcPremiumGem v-if="!isPremiumUser" />
      </div>
    </div>
    <div class="footer">
      <Button :disabled="publishDisabled" @click="share">
        {{ isNew ? $t('navigation.publish') : $t('navigation.update') }}
      </Button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.share-dashboard-modal-container {
  width: 360px;
  display: flex;
  flex-direction: column;
  gap: var(--padding-large);

  @media screen and (max-width: 400px) {
    width: unset;
  }

  .content {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--padding);

    .medium {
      font-weight: var(--standard_text_medium_font_weight);
    }

    .share-setting {
      display: flex;
      align-items: center;
      gap: var(--padding);
    }
  }

  .footer {
    display: flex;
    justify-content: center;
  }
}

:global(.share-dialog-setting-tooltip >div) {
  width: 190px;
}
</style>
