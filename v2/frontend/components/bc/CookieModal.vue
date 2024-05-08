<script lang="ts" setup>
import { COOKIE_KEY, type CookiesPreference } from '~/types/cookie'
import { Target } from '~/types/links'

const cookiePreference = useCookie<CookiesPreference>(COOKIE_KEY.COOKIES_PREFERENCE, { default: () => undefined })
const { t: $t } = useI18n()

const setCookiePreference = (value: CookiesPreference) => {
  cookiePreference.value = value
}

const visible = computed(() => cookiePreference.value === undefined)
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
      <div class="text-container">
        {{ tOf($t, 'cookies.text', 0) }}
        <NuxtLink to="https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf" :target="Target.External" class="link">
          {{ tOf($t, 'cookies.text', 1) }}
        </NuxtLink>
        {{ tOf($t, 'cookies.text', 2) }}
      </div>
      <div class="button-container">
        <Button class="necessary-button" @click="setCookiePreference('functional')">
          {{ $t('cookies.only_necessary') }}
        </Button>
        <Button @click="setCookiePreference('all')">
          {{ $t('cookies.accept_all') }}
        </Button>
      </div>
    </div>
  </Dialog>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.dialog-container {
  display: flex;
  align-items: center;
  gap: 35px;

  @media (max-width: 670px) {
    flex-direction: column;
    gap: var(--padding);
  }

  .text-container {
    @include fonts.standard_text;
    width: 320px;

    @media (max-width: 670px) {
      width: auto;
      max-width: 360px;
    }
  }

  .button-container {
    display: flex;
    gap: 7px;
    min-width: max-content;

    @media (max-width: 670px) {
      flex-direction: column;
      gap: 9px;
      width: 100%;
    }

    .necessary-button {
      background-color: var(--button-color-dark-pattern);
      border-color: var(--button-color-dark-pattern);
      color: var(--button-text-color-dark-pattern);
    }
  }
}
</style>
