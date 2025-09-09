<script setup lang="ts">
useWindowSizeProvider()
useBcToastProvider()

const { latestState } = storeToRefs(useLatestStateStore())
const exchangeRates = computed(() => latestState.value?.exchange_rates ?? [])
const exchangeRateLengthOnTestNetworks = 1
if (exchangeRates.value.length === exchangeRateLengthOnTestNetworks) {
  const { selectedCurrencyMain } = useCurrency()
  selectedCurrencyMain.value = exchangeRates.value[0]?.code as CurrencyCode
}

useHead({
  htmlAttrs: {
    class: 'validator-dashboard',
  },
})
</script>

<template>
  <div class="min-h-full">
    <NuxtLoadingIndicator color="var(--primary-color)" />
    <BcDataWrapper>
      <slot />
      <DynamicDialog />
      <Toast />
    </BcDataWrapper>
  </div>
</template>

<style lang="scss">
@use "~~/assets/css/main.scss";
@use "~~/assets/css/prime.scss";
</style>
