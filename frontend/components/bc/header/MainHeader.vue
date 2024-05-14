<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'

const props = defineProps({ isHomePage: { type: Boolean } })
const { latestState } = useLatestStateStore()
const { slotToEpoch } = useNetwork()
const { isLoggedIn } = useUserStore()
const { currency, available, rates } = useCurrency()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const loginText = computed(() => {
  return isLoggedIn.value ? 'Logged in' : 'Login'
})

const rate = computed(() => {
  if (isFiat(currency.value) && rates.value?.[currency.value]) {
    return rates.value[currency.value]
  } else if (rates.value?.USD) {
    return rates.value.USD
  }
  const fiat = available.value?.find(c => isFiat(c))
  if (fiat && rates.value?.[fiat]) {
    return rates.value[fiat]
  }
})

const currentEpoch = computed(() => latestState.value?.current_slot !== undefined ? slotToEpoch(latestState.value.current_slot) : undefined)

</script>

<template>
  <div class="header top">
    <div class="content">
      <div class="left-info">
        <span v-if="latestState?.current_slot"><span>{{ $t('main.current_slot') }}</span>:
          <NuxtLink :to="`/slot/${latestState.current_slot}`" class="bold" :no-prefetch="true">
            <BcFormatNumber :value="latestState.current_slot" />
          </NuxtLink>
        </span>
        <span v-if="currentEpoch !== undefined"><span>{{ $t('main.current_epoch') }}</span>:
          <NuxtLink :to="`/epoch/${currentEpoch}`" class="bold" :no-prefetch="true">
            <BcFormatNumber :value="currentEpoch" />
          </NuxtLink>
        </span>
        <span v-if="rate"><span><IconNetworkEthereum class="icon monochromatic" />ETH</span>:
          <span class="bold"> {{ rate.symbol }}<BcFormatNumber :value="rate.rate" :max-decimals="2" /></span>
        </span>
      </div>
      <BcSearchbarGeneral v-if="showInDevelopment && !props.isHomePage" bar-style="discreet" />
      <NuxtLink to="/login">
        {{ loginText }}
      </NuxtLink>
    </div>
  </div>
  <div class="header bottom">
    <div class="content">
      <NuxtLink to="/" class="logo">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </NuxtLink>

      <BcHeaderMegaMenu />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--light-grey);

  &.top {
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }

  &.bottom {
    min-height: var(--navbar2-height);
    background-color: var(--container-background);
    color: var(--container-color);
    border-bottom: 1px solid var(--container-border-color);
  }

  .content {
    width: var(--content-width);
    margin-left: var(--content-margin);
    margin-right: var(--content-margin);
    display: flex;
    align-items: flex-start;
    justify-content: space-between;

    .left-info {
      display: flex;
      gap: var(--padding-large);
      font-weight: var(--standard_text_light_font_weight);
      font-size: var(--title_font_size);

      .bold {
        font-weight: var(--standard_text_bold_font_weight);
      }

      .icon{
        height: 12px;
        width: auto;
        margin-right: var(--padding-small);
      }

    }
  }

  .logo {
    height: var(--navbar2-height);
    display: flex;
    align-items: center;
  }
}

.page {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>
