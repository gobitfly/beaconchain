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
    <div class="dialog-container">
      <div class="text-container">
        {{ $t('cookies.text') }}
      </div>
      <div class="button-container">
        <Button @click="setCookiePreference('all')">
          {{ $t('cookies.accept_all') }}
        </Button>
        <Button class="necessary-button" @click="setCookiePreference('functional')">
          {{ $t('cookies.only_necessary') }}
        </Button>
      </div>
    </div>
  </BcDialog>
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
    // TODO: Required? Check once font size PR is merged
    @include fonts.standard_text;
    //line-height: 1.5; // TODO: Required?
    width: 297px;

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
    }
  }
}

</style>
