<script lang="ts" setup>
import { warn } from 'vue'
import type { ValidatorDashboard } from '~/types/api/dashboard'
import { API_PATH } from '~/types/customFetch'
import { isSharedKey } from '~/utils/dashboard/key'

interface Props {
  // Currently only validator dashboards are supported. For public dashboards this will be undefined
  dashboard?: ValidatorDashboard
  dashboardKey: string
}
const {
  dialogRef, props,
} = useBcDialog<Props>()
const { t: $t } = useTranslation()
const router = useRouter()
const url = useRequestURL()
const { refreshDashboards } = useUserDashboardStore()
const { fetch } = useCustomFetch()
const { user } = useUserStore()

const isUpdating = ref(false)

const isReadonly = computed(() => !props.value?.dashboard)

const sharedKey = computed(() =>
  props.value?.dashboard
    ? props.value.dashboard.public_ids?.[0]?.public_id
    : props.value?.dashboardKey,
)

const isShared = computed(() => isSharedKey(sharedKey.value))

const path = computed(() => {
  const newRoute = router.resolve({
    name: 'dashboard-id',
    params: { id: sharedKey.value },
  })
  return url.origin + newRoute.fullPath
})

const edit = () => {
  if (isReadonly.value) {
    warn('cannot edit public dashboard share')
    return
  }
  dialogRef?.value?.close('EDIT')
}

const unpublish = async () => {
  if (isReadonly.value) {
    warn('cannot delete public dashboard share')
    return
  }
  if (isUpdating.value) {
    return
  }
  isUpdating.value = true
  const publicId = `${props.value?.dashboard?.public_ids?.[0]?.public_id}`
  await fetch(
    API_PATH.DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID,
    { method: 'DELETE' },
    {
      dashboardKey: `${props.value?.dashboard?.id}`,
      publicId,
    },
  )
  await refreshDashboards()
  dialogRef?.value?.close('DELETE')
  isUpdating.value = false
}
</script>

<template>
  <div class="share-dashboard-code-modal-container">
    <div class="content">
      <qrcode-vue
        class="qr-code"
        :value="path"
        :size="330"
      />
      <label class="title">{{
        $t("dashboard.share_dialog.public_dashboard_url")
      }}</label>
      <BcCopyLabel
        :value="path"
        class="copy_label"
      />
      <label
        v-if="isShared"
        class="disclaimer"
      >{{
        $t("dashboard.share_dialog.only_viewing_permission")
      }}</label>
      <label
        v-else
        class="disclaimer"
      >{{
        $t("dashboard.share_dialog.share_public_disclaimer")
      }}</label>
      <label
        v-if="!user?.premium_perks?.share_custom_dashboards"
        class="disclaimer"
      >{{ $t("dashboard.share_dialog.upgrade") }}<BcPremiumGem class="gem" /></label>
      <div
        v-if="!isReadonly"
        class="footer"
      >
        <Button
          :disabled="isUpdating"
          @click="unpublish"
        >
          {{ $t("navigation.unpublish") }}
        </Button>
        <Button
          :disabled="isUpdating"
          @click="edit"
        >
          {{ $t("dashboard.share_dialog.edit") }}
        </Button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.share-dashboard-code-modal-container {
  width: 335px;
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

    .disclaimer {
      font-size: var(--small_text_font_size);
      font-weight: var(--standard_text_light_font_weight);
    }

    .gem {
      display: inline-block;
    }

    .qr-code {
      border: 5px solid white;
    }

    .copy_label {
      width: 100%;
    }

    .title {
      font-size: var(--small_text_font_size);
      font-weight: var(--small_text_bold_font_weight);
    }
  }

  .footer {
    width: 100%;
    display: flex;
    justify-content: center;
    gap: var(--padding);
  }
}
</style>
