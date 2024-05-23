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
  <div class="anchor">
    <div class="top-background" />
    <div class="rows">
      <div class="grid-cell blockchain-info">
        <span v-if="latestState?.current_slot"><span>{{ $t('header.current_slot') }}</span>:
          <NuxtLink :to="`/slot/${latestState.current_slot}`" :no-prefetch="true" :disabled="!showInDevelopment || null">
            <BcFormatNumber class="bold" :value="latestState.current_slot" />
          </NuxtLink>
        </span>
        <span v-if="currentEpoch !== undefined"><span>{{ $t('header.current_epoch') }}</span>:
          <NuxtLink :to="`/epoch/${currentEpoch}`" :no-prefetch="true" :disabled="!showInDevelopment || null">
            <BcFormatNumber class="bold" :value="currentEpoch" />
          </NuxtLink>
        </span>
        <span v-if="rate">
          <span>
            <IconNetworkEthereum class="network-icon monochromatic" />ETH
          </span>:
          <span> {{ rate.symbol }}
            <BcFormatNumber class="bold" :value="rate.rate" :max-decimals="2" />
          </span>
        </span>
      </div>

      <div class="grid-cell search-bar">
        <BcSearchbarGeneral v-if="showInDevelopment && !props.isHomePage" class="bar" :bar-style="width < 1023 ? 'gaudy' : 'discreet'" />
      </div>

      <div class="grid-cell controls">
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

      <div class="grid-cell logo">
        <NuxtLink to="/" class="logo-component">
          <IconBeaconchainLogo alt="Beaconcha.in logo" />
          beaconcha.in
        </NuxtLink>
      </div>

      <div class="grid-cell mega-menu">
        <BcHeaderMegaMenu ref="megaMenu" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.anchor {
  position: relative;
  display: flex;
  width: 100%;
  justify-content: center;
  border-bottom: 1px solid var(--container-border-color);
  background-color: var(--container-background);
  color: var(--container-color);
  .top-background {
    position: absolute;
    width: 100%;
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }

  .rows {
    position: relative;
    display: grid;
    grid-template-columns: min-content min-content auto min-content;
    grid-template-rows: var(--navbar-height) min-content;
    @media (max-width: 1023px) {
      grid-template-columns: min-content auto min-content;
      grid-template-rows: var(--navbar-height) min-content;
    }
    width: var(--content-width);
    margin-left: var(--content-margin);
    margin-right: var(--content-margin);
    color: var(--light-grey);
    @mixin bottom-cell($row) {
      color: var(--container-color);
      grid-row: $row;
    }
    font-family: var(--main_header_font_family);
    font-size: var(--main_header_font_size);
    font-weight: var(--main_header_font_weight);
    .bold {
      font-weight: var(--main_header_bold_font_weight);
    }
    .grid-cell {
      position: relative;
      display: flex;
      margin-top: auto;
      margin-bottom: auto;
      align-items: center;
      vertical-align: middle;
      height: 100%;
      gap: var(--padding);
    }

    .blockchain-info {
      @media (min-width: 1023.1px) {
        grid-row: 1;
        grid-column: 1;
        grid-column-end: span 2;
      }
      @media (max-width: 1023px) {
        display: none;
      }
      white-space: nowrap;
      .network-icon {
        height: 14px;
        width: auto;
        margin-right: var(--padding-small);
      }
    }

    .search-bar {
      grid-row: 1;
      grid-column: 3;
      @media (max-width: 1023px) {
        @include bottom-cell(3);
        grid-column: 1;
        grid-column-end: span 3;
      }
      .bar {
        position: relative;
        margin-top: var(--content-margin);
        margin-bottom: var(--content-margin);
        @media (max-width: 1023px) {
          width: 100%;
        }
      }
    }

    .controls {
      grid-row: 1;
      grid-column: 4;
      @media (max-width: 1023px) {
        grid-column: 3;
      }
      .currency {
        @media (max-width: 469px) {
          display: none;
        }
      }
      .logged-out {
        white-space: nowrap;
        display: flex;
        align-items: center;
        gap: var(--padding-small);
        .signup {
          padding: 8px;
        }
      }
      .user-menu-icon {
        width: 19px;
        height: 18px;
        color: var(--light-grey);
      }
      .burger {
        cursor: pointer;
      }
    }

    .logo {
      grid-column: 1;
      @media (min-width: 1023.1px) {
        @include bottom-cell(2);
      }
      @media (max-width: 1023px) {
        grid-row: 1;
      }
      .logo-component {
        display: flex;
        align-items: flex-end;
        gap: var(--padding);
        font-family: var(--logo_font_family);
        font-size: var(--logo_font_size);
        font-weight: var(--logo_font_weight);
        letter-spacing: var(--logo_letter_spacing);
        line-height: 20px;
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
    }

    .mega-menu {
      @media (min-width: 1023.1px) {
        grid-column: 2;
        grid-column-end: span 3;
        @include bottom-cell(2);
      }
      @media (max-width: 1023px) {
        grid-row: 2;
        grid-column: 1;
        grid-column-end: span 3;
      }
    }
  }
}
</style>
