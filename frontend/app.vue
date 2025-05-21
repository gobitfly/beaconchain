<script setup lang="ts">
// https://i18n.nuxtjs.org/docs/guide/seo#setup
// useLocaleHead() did not work
const { locale } = useTranslation()
useHead(
  {
    htmlAttrs: { lang: locale.value },
  },
)
useWindowSizeProvider()
useBcToastProvider()

const { latestState } = storeToRefs(useLatestStateStore())
const exchangeRates = computed(() => latestState.value?.exchange_rates ?? [])
const exchangeRateLengthOnTestNetworks = 1
if (exchangeRates.value.length === exchangeRateLengthOnTestNetworks) {
  const { selectedCurrencyMain } = useCurrency()
  selectedCurrencyMain.value = exchangeRates.value[0].code as CurrencyCode
}
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
