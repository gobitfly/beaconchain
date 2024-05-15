<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faBars
} from '@fortawesome/pro-solid-svg-icons'
import type { BcHeaderMegaMenu } from '#build/components'
import { useLatestStateStore } from '~/stores/useLatestStateStore'

const props = defineProps({ isHomePage: { type: Boolean } })
const { latestState } = useLatestStateStore()
const { slotToEpoch } = useNetwork()
const { doLogout } = useUserStore()
const { isLoggedIn } = useUserStore()
const { currency, available, rates } = useCurrency()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const megaMenu = ref<typeof BcHeaderMegaMenu | null>(null)

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

const toggleMegaMenu = (evt: Event) => {
  megaMenu.value?.toggleMegaMenu(evt)
}

</script>

<template>
  <div class="header top">
    <div class="content">
      <div class="left-content">
        <span v-if="latestState?.current_slot"><span>{{ $t('header.current_slot') }}</span>:
          <NuxtLink :to="`/slot/${latestState.current_slot}`" :no-prefetch="true">
            <BcFormatNumber :value="latestState.current_slot" />
          </NuxtLink>
        </span>
        <span v-if="currentEpoch !== undefined"><span>{{ $t('header.current_epoch') }}</span>:
          <NuxtLink :to="`/epoch/${currentEpoch}`" :no-prefetch="true">
            <BcFormatNumber :value="currentEpoch" />
          </NuxtLink>
        </span>
        <span v-if="rate"><span>
          <IconNetworkEthereum class="icon monochromatic" />ETH
        </span>:
          <span> {{ rate.symbol }}
            <BcFormatNumber :value="rate.rate" :max-decimals="2" />
          </span>
        </span>
      </div>
      <BcSearchbarGeneral v-if="showInDevelopment && !props.isHomePage" bar-style="discreet" />
      <div class="right-content">
        <BcCurrencySelection />
        <NuxtLink v-if="!isLoggedIn" to="/login">
          {{ $t('header.login') }}
        </NuxtLink>
        <div v-else @click="doLogout">
          logout
        </div>
        <div class="burger" @click.stop.prevent="toggleMegaMenu">
          <FontAwesomeIcon :icon="faBars" />
        </div>
      </div>
    </div>
  </div>
  <div class="header bottom">
    <div class="content">
      <NuxtLink to="/" class="logo">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </NuxtLink>

      <BcHeaderMegaMenu ref="megaMenu" />
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

    .content {
      align-items: center;
    }
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
    justify-content: space-between;
    font-family: var(--main_header_font_size);
    font-size: var(--main_header_font_size);
    font-weight: var(--main_header_font_weight);

    .left-content {
      display: flex;
      align-items: center;
      gap: var(--padding-large);

      .icon {
        height: 14px;
        width: auto;
        margin-right: var(--padding-small);
      }

    }

    .right-content {
      display: flex;
      align-items: center;
      gap: var(--padding-large);
    }
  }

  .logo {
    height: var(--navbar2-height);
    display: flex;
    align-items: center;
  }

  .burger {
    cursor: pointer;
  }
}

.page {
  display: flex;
  flex-direction: column;
  align-items: center;
}
</style>
