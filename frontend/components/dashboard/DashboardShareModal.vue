<script lang="ts" setup>
import type { ValidatorDashboard } from '~/types/api/dashboard'

const {
  dialogRef,
  props,
} = useBcDialog<{
  dashboard?: ValidatorDashboard,
}>()
const { t: $t } = useTranslation()
// const { refresh } = useUserDashboardStore()
// const { fetch } = useCustomFetch()
const {
  name,
  publicName,
} = useDashboard()
const dashboardName = ref(publicName.value ?? name.value ?? '')
const shareGroups = ref(true)
const isUpdating = ref(false)

const { hasShareCustomDashboard } = usePremiumPerks()

const { $api } = useNuxtApp()
const { refresh } = usePrivateDashboards()
const add = async () => {
  isUpdating.value = true
  await $api(
    `/api/bff/validator-dashboards/${props.value?.dashboard?.id}/public-ids`,
    {
      body: {
        name: dashboardName.value,
        share_settings: { share_groups: shareGroups.value },
      },
      method: 'post',
    },
  )
  await refresh()
  dialogRef?.value?.close(true)
  isUpdating.value = false
}

const edit = async () => {
  isUpdating.value = true
  const publicId = `${props.value?.dashboard?.public_ids?.[0]?.public_id}`
  await $api(
    `/api/bff/validator-dashboards/${props.value?.dashboard?.id}/public-ids/${publicId}`,
    {
      body: {
        name: dashboardName.value,
        share_settings: { share_groups: shareGroups.value },
      },
      method: 'put',
    },
  )
  await refresh()
  dialogRef?.value?.close(true)
  isUpdating.value = false
}

const publishDisabled = computed(() => {
  return isUpdating.value || !REGEXP_VALID_NAME.test(dashboardName.value)
})

const share = () => {
  if (publishDisabled.value) {
    return
  }

  if (props.value?.dashboard?.public_ids?.[0]?.public_id) {
    edit()
  }
  else {
    add()
  }
}

const shareGroupTooltip = computed(() => {
  return formatMultiPartSpan(
    $t,
    'dashboard.share_dialog.setting.group.tooltip',
    [
      undefined,
      'bold',
      undefined,
    ],
  )
})
</script>

<template>
  <div class="share-dashboard-modal-container">
    <div class="content">
      <label
        for="dashboardName"
        class="medium"
      >{{
        $t("dashboard.share_dialog.setting.name.label")
      }}</label>
      <InputText
        id="dashboardName"
        v-model.trim="dashboardName"
        class="input-field"
        @keypress.enter="share"
      />
      <div class="share-setting">
        <Checkbox
          v-model="shareGroups"
          input-id="shareGroup"
          :binary="true"
          :disabled="!hasShareCustomDashboard"
        />
        <label
          for="shareGroup"
          :class="{ 'text-disabled': !hasShareCustomDashboard }"
        >{{
          $t("dashboard.share_dialog.setting.group.label")
        }}</label>

        <BcTooltip
          position="top"
          tooltip-class="share-dialog-setting-tooltip"
          :text="shareGroupTooltip"
          :render-text-as-html="true"
        >
          <BcIcon name="circle-info" />
        </BcTooltip>
        <BcPremiumGem v-if="!hasShareCustomDashboard" />
      </div>
    </div>
    <div class="footer">
      <Button
        :disabled="publishDisabled"
        @click="share"
      >
        {{ publicName ? $t("navigation.update") : $t("navigation.publish") }}
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

:global(.share-dialog-setting-tooltip > div) {
  width: 190px;
}
</style>
