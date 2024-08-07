<script lang="ts" setup>
import { COOKIE_KEY, type CookiesPreference } from '~/types/cookie'

const cookiePreference = useCookie<CookiesPreference>(COOKIE_KEY.COOKIES_PREFERENCE, { default: () => undefined })
const { isShared } = useDashboardKey()
const { dashboards } = useUserDashboardStore()
const { t: $t } = useTranslation()
const route = useRoute()

const dismissed = ref(false)
const visible = computed(() => isShared.value && !dismissed.value && cookiePreference.value !== undefined)

const text = computed(() => {
  const userHasOwnDashboard = ((dashboards.value?.validator_dashboards?.length || 0) + (dashboards.value?.account_dashboards?.length || 0)) > 0
  const textRoot = userHasOwnDashboard ? 'dashboard.shared_modal_with_own' : 'dashboard.shared_modal_without_own'

  const caption = $t(textRoot + '.text')
  const button = $t(textRoot + '.button')

  return { caption, button }
})
</script>

<template>
  <Dialog
    v-model:visible="visible"
    :dismissable-mask="false"
    :draggable="false"
    :close-on-escape="false"
    position="bottom"
  >
    <div class="dialog-container">
      {{ text.caption }}
      <div class="button-row">
        <div class="dismiss" @click="dismissed=true">
          {{ $t('navigation.dismiss') }}
        </div>
        <BcLink :to="`/dashboard`" :replace="route.path.startsWith('/dashboard')">
          <Button>
            {{ text.button }}
          </Button>
        </BcLink>
      </div>
    </div>
  </Dialog>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.dialog-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--padding);

  .button-row {
    display: flex;
    align-items: center;
    gap: var(--padding-large);

    .dismiss {
      cursor: pointer;
      color: var(--text-color-disabled);
    }
  }
}
</style>
