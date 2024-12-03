<script lang="ts" setup>
import {
  COOKIE_KEY, type CookiesPreference,
} from '~/types/cookie'

const cookiePreference = useCookie<CookiesPreference>(
  COOKIE_KEY.COOKIES_PREFERENCE,
  {
    default: () => undefined,
    maxAge: 60 * 60 * 24 * 365,
  },
)
const { t: $t } = useTranslation()

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
      <BcTranslation
        class="text-container"
        keypath="cookies.text.template"
        linkpath="cookies.text._link"
        to="https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf"
      />
      <div class="button-container">
        <div
          class="necessary-button"
          @click="setCookiePreference('functional')"
        >
          {{ $t("cookies.only_necessary") }}
        </div>
        <Button @click="setCookiePreference('all')">
          {{ $t("cookies.accept_all") }}
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
    align-items: center;
    gap: var(--padding-large);
    min-width: max-content;

    @media (max-width: 670px) {
      flex-direction: column;
      gap: 9px;
      width: 100%;

      > Button {
        width: 100%;
      }
    }

    .necessary-button {
      cursor: pointer;
      color: var(--text-color-disabled);
    }
  }
}
</style>
