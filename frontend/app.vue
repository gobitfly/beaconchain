<script setup lang="ts">
// https://i18n.nuxtjs.org/docs/guide/seo#setup
// useLocaleHead() did not work
const { locale } = useTranslation()
useHead(
  {
    htmlAttrs: { lang: locale.value },
  },
)
useBcToastProvider()
</script>

<template>
  <NuxtLoadingIndicator color="var(--primary-color)" />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <DynamicDialog />
  <Toast />
  <BcCookieModal />
</template>

<style lang="scss"></style>
