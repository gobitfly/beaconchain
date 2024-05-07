<script lang="ts" setup>
import { COOKIE_KEY, type CookiesPreference } from '~/types/cookie'

const cookiePreference = useCookie<CookiesPreference>(COOKIE_KEY.COOKIES_PREFERENCE, { default: () => undefined })
const { t: $t } = useI18n()

const setCookiePreference = (value: CookiesPreference) => {
  cookiePreference.value = value
}

const visible = computed(() => cookiePreference.value === undefined)
const modalText = computed(() => formatMultiPartSpan($t, 'cookies.text', [undefined, 'link', undefined], [undefined, 'https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf', undefined]))

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
      <!--eslint-disable-next-line vue/no-v-html-->
      <div class="text-container" v-html="modalText" />
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
