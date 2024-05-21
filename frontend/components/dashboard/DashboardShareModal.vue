<script lang="ts" setup>
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { ValidatorDashboard } from '~/types/api/dashboard'

interface Props {
  dashboard: ValidatorDashboard; // Currently only validator dashboards are supported
}
const { props, setHeader } = useBcDialog<Props>()
const { t: $t } = useI18n()

const dashboardName = ref('')
const shareGroups = ref(true)

watch(props, (p) => {
  if (p) {
    setHeader(p.dashboard.name)
    dashboardName.value = p.dashboard.public_ids?.[0]?.name ?? ''
  }
}, { immediate: true })

const share = () => {
  // TODO
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
    <Button @click="share">
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
