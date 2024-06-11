<script lang="ts" setup>
import { COOKIE_KEY, type CookiesPreference } from '~/types/cookie'

const cookiePreference = useCookie<CookiesPreference>(COOKIE_KEY.COOKIES_PREFERENCE, { default: () => undefined })
const { isShared } = useDashboardKey()
const { t: $t } = useI18n()
const route = useRoute()

const visible = computed(() => isShared.value && cookiePreference.value !== undefined)
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
      {{ $t('dashboard.shared_modal.text') }}
      <BcLink :to="`/dashboard`" :replace="route.path.startsWith('/dashboard')">
        <Button class="get-started">
          {{ $t('dashboard.shared_modal.get_started') }}
        </Button>
      </BcLink>
    </div>
  </Dialog>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.dialog-container {
  display: flex;
  align-items: center;
  gap: var(--padding-large);

  .get-started {
    min-width: 120px;
  }
}
</style>
