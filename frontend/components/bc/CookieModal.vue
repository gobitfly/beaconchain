<script lang="ts" setup>
import { COOKIE_KEY, type CookiesPreference } from '~/types/cookie'

const cookiePreference = useCookie<CookiesPreference>(COOKIE_KEY.COOKIES_PREFERENCE, { default: () => undefined })
const { t: $t } = useI18n()

const setCookiePreference = (value: CookiesPreference) => {
  cookiePreference.value = value
}

const visible = computed(() => cookiePreference.value === undefined)

</script>

<template>
  <BcDialog v-model="visible">
    {{ $t('cookies.text') }}
    <Button @click="setCookiePreference('all')">
      {{ $t('cookies.accept_all') }}
    </Button>
    <Button @click="setCookiePreference('functional')">
      {{ $t('cookies.only_necessary') }}
    </Button>
  </BcDialog>
</template>

<style lang="scss" scoped>

</style>
