<script lang="ts" setup>
import { warn } from 'vue'
import type { ValidatorDashboard } from '~/types/api/dashboard'

// import { isSharedDashboardKey } from '~/utils/dashboard/key'

const {
  dialogRef,
  props,
} = useBcDialog<{
  // Currently only validator dashboards are supported. For guest dashboards this will be undefined
  dashboard?: ValidatorDashboard,
  dashboardKey: string,
}>()
const { t: $t } = useTranslation()
const {
  publicId,
  variant,
} = useDashboard()
const {
  currentUrl,
  origin,
} = useUrl()
const url = variant.value === 'guest-dashboard'
  ? currentUrl
  : `${origin}/dashboard/${publicId.value}`
// const { refresh } = useUserDashboardStore()
// const { fetch } = useCustomFetch()
// const { user } = useUserStore()

const isUpdating = ref(false)

const isReadonly = computed(() => !props.value?.dashboard)

// const sharedKey = computed(() =>
//   props.value?.dashboard
//     ? props.value.dashboard.public_ids?.[0]?.public_id
//     : props.value?.dashboardKey,
// )

// const isShared = computed(() => isSharedDashboardKey(sharedKey.value))

// const path = computed(() => {
//   const newRoute = router.resolve({
//     name: 'dashboard-id',
//     params: { id: sharedKey.value },
//   })
//   return url.origin + newRoute.fullPath
// })

const edit = () => {
  if (isReadonly.value) {
    warn('cannot edit guest dashboard share')
    return
  }
  dialogRef?.value?.close('EDIT')
}
const { $api } = useNuxtApp()
const { refresh } = usePrivateDashboards()
const unpublish = async () => {
  if (isReadonly.value) {
    warn('cannot delete guest dashboard share')
    return
  }
  if (isUpdating.value) {
    return
  }
  isUpdating.value = true
  const publicId = `${props.value?.dashboard?.public_ids?.[0]?.public_id}`
  await $api(`/api/bff/validator-dashboards/${props.value?.dashboard?.id}/public-ids/${publicId}`,
    { method: 'delete' },
  )
  await refresh()
  dialogRef?.value?.close('DELETE')
  isUpdating.value = false
}

const { hasShareCustomDashboard } = usePremiumPerks()
</script>

<template>
  <div class="share-dashboard-code-modal-container">
    <div class="content">
      <Qrcode
        class="qr-code"
        variant="rounded"
        :value="url"
      />
      <label class="title">{{
        $t("dashboard.share_dialog.public_dashboard_url")
      }}</label>
      <BcCopyLabel
        :value="url"
        class="copy_label"
      />
      <p
        class="disclaimer"
      >
        {{
          variant === 'guest-dashboard'
            ? $t("dashboard.share_dialog.share_public_disclaimer")
            : $t("dashboard.share_dialog.only_viewing_permission")
        }}
      </p>
      <label
        v-if="!hasShareCustomDashboard"
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
      border-radius: var(--border-radius);
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
