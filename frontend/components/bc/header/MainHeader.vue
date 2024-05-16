<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faBars,
  faCircleUser
} from '@fortawesome/pro-solid-svg-icons'
import type { BcHeaderMegaMenu } from '#build/components'
import { useLatestStateStore } from '~/stores/useLatestStateStore'

const props = defineProps({ isHomePage: { type: Boolean } })
const { latestState } = useLatestStateStore()
const { slotToEpoch } = useNetwork()
const { doLogout, isLoggedIn } = useUserStore()
const { currency, available, rates } = useCurrency()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const { width } = useWindowSize()
const { t: $t } = useI18n()

const isSmallScreen = computed(() => width.value <= 1023)
const isMobile = computed(() => width.value <= 469)

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

const userMenu = computed(() => {
  return [
    {
      label: $t('header.logout'),
      command: () => doLogout()
    }
  ]
})

</script>

<template>
  <div class="header top">
    <div class="content">
      <div class="left-content">
        <NuxtLink to="/dashboard" class="logo">
          <IconBeaconchainLogo alt="Beaconcha.in logo" />
          beaconcha.in
        </NuxtLink>
        <span v-if="latestState?.current_slot" class="info"><span>{{ $t('header.current_slot') }}</span>:
          <NuxtLink :to="`/slot/${latestState.current_slot}`" :no-prefetch="true" :disabled="!showInDevelopment">
            <BcFormatNumber :value="latestState.current_slot" />
          </NuxtLink>
        </span>
        <span v-if="currentEpoch !== undefined" class="info"><span>{{ $t('header.current_epoch') }}</span>:
          <NuxtLink :to="`/epoch/${currentEpoch}`" :no-prefetch="true" :disabled="!showInDevelopment">
            <BcFormatNumber :value="currentEpoch" />
          </NuxtLink>
        </span>
        <span v-if="rate" class="info">
          <span>
            <IconNetworkEthereum class="icon monochromatic" />ETH
          </span>:
          <span> {{ rate.symbol }}
            <BcFormatNumber :value="rate.rate" :max-decimals="2" />
          </span>
        </span>
      </div>
      <BcSearchbarGeneral v-if="showInDevelopment && !props.isHomePage" class="search" bar-style="discreet" />
      <div class="right-content">
        <BcCurrencySelection v-if="!isMobile" class="currency" />
        <div v-if="!isLoggedIn" class="logged-out">
          <NuxtLink to="/login">
            {{ $t('header.login') }}
          </NuxtLink>
          /
          <NuxtLink to="/signup">
            <Button class="signup" :label="$t('header.signup')" />
          </NuxtLink>
        </div>
        <div v-else-if="!isSmallScreen">
          <BcDropdown :options="userMenu" variant="header" option-label="label">
            <template #value>
              <FontAwesomeIcon class="user-menu-icon" :icon="faCircleUser" />
            </template>
            <template #option="slotProps">
              <span @click="slotProps.command?.()">
                {{ slotProps.label }}
              </span>
            </template>
          </BcDropdown>
        </div>
        <div v-if="isSmallScreen" class="burger" @click.stop.prevent="toggleMegaMenu">
          <FontAwesomeIcon :icon="faBars" />
        </div>
      </div>
    </div>
  </div>
  <div class="header bottom">
    <div class="content">
      <NuxtLink to="/dashboard" class="logo">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
        beaconcha.in
      </NuxtLink>
      <BcHeaderMegaMenu ref="megaMenu" />
      <BcSearchbarGeneral v-if="showInDevelopment && !props.isHomePage" class="search" bar-style="discreet" />
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

    .user-menu-icon {
      width: 19px;
      height: 18px;
      color: var(--light-grey);
    }

    .logo {
      display: none;
    }

    @media (max-width: 1023px) {

      .search,
      .info {
        display: none;
      }

      .logo {
        display: flex;
      }
    }

    @media (max-width: 469px) {
      .currency {
        display: none;
      }
    }
  }

  &.bottom {
    min-height: var(--navbar2-height);
    background-color: var(--container-background);
    color: var(--container-color);
    border-bottom: 1px solid var(--container-border-color);

    .search {
      display: none;
    }

    @media (max-width: 1023px) {
      min-height: unset;

      .content {
        flex-direction: column;
      }

      .logo {
        display: none;
      }

      .search {
        display: flex;
        width: 100%;
        margin-top: var(--content-margin);
        margin-bottom: var(--content-margin);
      }
    }
  }

  .content {
    width: var(--content-width);
    margin-left: var(--content-margin);
    margin-right: var(--content-margin);
    align-items: center;
    display: flex;
    justify-content: space-between;
    font-family: var(--main_header_font_family);
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

      .logged-out {
        white-space: nowrap;
        display: flex;
        align-items: center;
        gap: var(--padding-small);

        .signup {
          padding: 8px;
        }
      }
    }
  }

  .logo {
    display: flex;
    align-items: flex-end;
    gap: var(--padding);
    font-family: var(--logo_font_family);
    font-size: var(--logo_font_size);
    font-weight: var(--logo_font_weight);
    letter-spacing: var(--logo_letter_spacing);
    line-height: 18px;

    @media (max-width: 1359px) {
      font-size: var(--logo_small_font_size);
      letter-spacing: var(--logo_small_letter_spacing);
      gap: 6px;
      align-items: center;

      svg {
        height: 18px;
        margin-bottom: 7px;
      }
    }

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
