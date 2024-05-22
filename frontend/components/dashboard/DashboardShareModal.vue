<script lang="ts" setup>
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { warn } from 'vue'
import type { ValidatorDashboard } from '~/types/api/dashboard'
import { API_PATH } from '~/types/customFetch'

interface Props {
  dashboard: ValidatorDashboard; // Currently only validator dashboards are supported
}
const { props, setHeader } = useBcDialog<Props>()
const { t: $t } = useI18n()
const { refreshDashboards } = useUserDashboardStore()
const { fetch } = useCustomFetch()

const dashboardName = ref('')
const shareGroups = ref(true)
const isUpdating = ref(false)

watch(props, (p) => {
  if (p) {
    setHeader(p.dashboard.name)
    dashboardName.value = p.dashboard.public_ids?.[0]?.name ?? ''
    shareGroups.value = !!p.dashboard.public_ids?.[0]?.share_settings.group_names
  }
}, { immediate: true })

const add = async () => {
  if (isUpdating.value) {
    return
  }
  warn('props.value?.dashboard', props.value?.dashboard)
  isUpdating.value = true
  await fetch(API_PATH.DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID, { body: { name: dashboardName.value, share_settings: { group_names: shareGroups.value } } }, { dashboardKey: `${props.value?.dashboard.id}` })
  await refreshDashboards()
  isUpdating.value = false
}

const edit = async () => {
  isUpdating.value = true
  await fetch(API_PATH.DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID, { body: { name: dashboardName.value, share_settings: { group_names: shareGroups.value } } }, { dashboardKey: `${props.value?.dashboard.id}`, publicId: `${props.value?.dashboard.public_ids?.[0]?.public_id}` })
  await refreshDashboards()

  isUpdating.value = false
}

const share = () => {
  if (props.value?.dashboard.public_ids?.[0]?.public_id) {
    edit()
  } else {
    add()
  }
}

const shareGroupTooltip = computed(() => {
  return formatMultiPartSpan($t, 'dashboard.share.setting.group.tooltip', [undefined, 'bold', undefined])
})

</script>

<template>
  <div class="share-dashboard-modal-container">
    <div class="content">
      <InputText v-model="dashboardName" :placeholder="$t('dashboard.share.placeholder')" class="input-field" />
      <div class="share-setting">
        <Checkbox id="shareGroup" v-model="shareGroups" :binary="true" />
        <label for="shareGroup">{{ $t('dashboard.share.setting.group.label') }}</label>

        <BcTooltip position="top" :text="shareGroupTooltip" :render-text-as-html="true">
          <FontAwesomeIcon :icon="faInfoCircle" />
        </BcTooltip>
        <BcPremiumGem /><!--TODO: only show gem for free users once we have that information-->
      </div>
    </div>
    <Button :disabled="isUpdating" @click="share">
      {{ $t('dashboard.share.share') }}
    </Button>
  </div>
</template>

<style lang="scss" scoped>
.share-dashboard-modal-container {
  width: 410px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--padding-large);

  @media screen and (max-width: 500px) {
    width: unset;
    height: unset;
  }

  .content {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--padding);

    .share-setting {
      display: flex;
      align-items: center;
      gap: var(--padding);
    }
  }
}
</style>
