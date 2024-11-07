<script setup lang="ts">
// https://i18n.nuxtjs.org/docs/guide/seo#setup
// useLocaleHead() did not work
const { locale } = useTranslation()
useHead(
  {
    htmlAttrs: { lang: locale.value },
    script: [ {
      async: false,
      key: 'revive',
      src: '../js/revive.min.js',
    } ],
  },
  { mode: 'client' },
)
useWindowSizeProvider()
useBcToastProvider()
useDateProvider()
</script>

<template>
  <div class="min-h-full">
    <BcDataWrapper>
      <NuxtLoadingIndicator color="var(--primary-color)" />
      <NuxtPage />
      <DynamicDialog />
      <Toast />
    </BcDataWrapper>
  </div>
</template>

<style lang="scss"></style>
