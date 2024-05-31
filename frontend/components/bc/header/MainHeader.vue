<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faBars,
  faCircleUser
} from '@fortawesome/pro-solid-svg-icons'
import type { BcHeaderMegaMenu } from '#build/components'
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { SearchbarShape, SearchbarColors } from '~/types/searchbar'
import { smallHeaderThreshold } from '~/types/header'

const props = defineProps({ isHomePage: { type: Boolean } })
const { latestState } = useLatestStateStore()
const { slotToEpoch } = useNetwork()
const { doLogout, isLoggedIn } = useUserStore()
const { currency, available, rates } = useCurrency()
const { width } = useWindowSize()
const { t: $t } = useI18n()

const colorMode = useColorMode()
const isSmallScreen = computed(() => width.value < smallHeaderThreshold)

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const hideInDevelopmentClass = showInDevelopment ? '' : 'hide-because-it-is-unfinished'

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

const isMobileMegaMenuOpen = computed(() => megaMenu.value?.isMobileMenuOpen)

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
  <div class="anchor" :class="hideInDevelopmentClass">
    <div class="top-background" />
    <div class="rows">
      <div class="grid-cell blockchain-info">
        <span v-if="latestState?.current_slot"><span>{{ $t('header.current_slot') }}</span>:
          <BcLink :to="`/slot/${latestState.current_slot}`" :disabled="!showInDevelopment || null">
            <BcFormatNumber class="bold" :value="latestState.current_slot" />
          </BcLink>
        </span>
        <span v-if="currentEpoch !== undefined"><span>{{ $t('header.current_epoch') }}</span>:
          <BcLink :to="`/epoch/${currentEpoch}`" :disabled="!showInDevelopment || null">
            <BcFormatNumber class="bold" :value="currentEpoch" />
          </BcLink>
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
        <BcSearchbarGeneral
          v-if="showInDevelopment && !props.isHomePage"
          class="bar"
          :bar-shape="SearchbarShape.Medium"
          :color-theme="isSmallScreen && colorMode.value != 'dark' ? SearchbarColors.LightBlue : SearchbarColors.DarkBlue"
          :screen-width-causing-sudden-change="smallHeaderThreshold"
        />
      </div>

      <div class="grid-cell controls">
        <BcCurrencySelection class="currency" />
        <div v-if="!isLoggedIn" class="logged-out">
          <BcLink to="/login" class="login">
            {{ $t('header.login') }}
          </BcLink>
          |
          <BcLink to="/register">
            <Button class="register" :label="$t('header.register')" />
          </BcLink>
        </div>
        <div v-else-if="!isSmallScreen" class="user-menu">
          <BcDropdown :options="userMenu" variant="header" option-label="label" class="menu-component">
            <template #value>
              <FontAwesomeIcon class="menu-icon" :icon="faCircleUser" />
            </template>
            <template #option="slotProps">
              <span @click="slotProps.command?.()">
                {{ slotProps.label }}
              </span>
            </template>
          </BcDropdown>
        </div>
        <FontAwesomeIcon :icon="faBars" class="burger" @click.stop.prevent="toggleMegaMenu" />
      </div>

      <div class="grid-cell logo">
        <BcLink to="/" class="logo-component">
          <IconBeaconchainLogo alt="Beaconcha.in logo" />
          beaconcha.in
        </BcLink>
      </div>

      <div class="grid-cell mega-menu">
        <BcHeaderMegaMenu ref="megaMenu" />
        <div v-if="isMobileMegaMenuOpen" class="decoration" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

// do not change these two values without changing the values in types/header.ts accordingly
$mobileHeaderThreshold: 470px;
$smallHeaderThreshold: 1024px;

.anchor {
  top: -1px;
  position: relative;
  display: flex;
  width: 100%;
  justify-content: center;
  border-bottom: 1px solid var(--container-border-color);
  &.hide-because-it-is-unfinished {
    border-bottom: none;
  }
  background-color: var(--container-background);
  .top-background {
    position: absolute;
    width: 100%;
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }

  .rows {
    position: relative;
    display: grid;
    grid-template-columns: 0px min-content min-content auto min-content 0px;  // the 0px are paddings, useless now but they exist in the structure of the grid so ready to be set if they are wanted one day
    grid-template-rows: var(--navbar-height) min-content;
    @media (max-width: $smallHeaderThreshold) {
      grid-template-columns: 0px min-content auto min-content 0px;  // same remark about the 0px
      grid-template-rows: var(--navbar-height) min-content;
    }
    width: var(--content-width);
    color: var(--header-top-font-color);
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
      flex-wrap: nowrap;
      white-space: nowrap;
      gap: var(--padding);
    }

    .blockchain-info {
      @media (min-width: $smallHeaderThreshold) {
        grid-row: 1;
        grid-column: 2;
        grid-column-end: span 2;
      }
      @media (max-width: $smallHeaderThreshold) {
        display: none;
      }
      margin-right: var(--padding-large);
      .network-icon {
        height: 14px;
        width: auto;
        margin-right: var(--padding-small);
      }
    }

    .search-bar {
      grid-row: 1;
      grid-column: 4;
      @media (max-width: $smallHeaderThreshold) {
        @include bottom-cell(3);
        grid-column: 2;
        grid-column-end: span 3;
      }
      .bar {
        position: relative;
        width: 100%;
        @media (min-width: $smallHeaderThreshold) {
          max-width: 460px;
        }
        margin-top: var(--content-margin);
        margin-bottom: var(--content-margin);
      }
    }

    .controls {
      user-select: none;
      grid-row: 1;
      grid-column: 5;
      @media (max-width: $smallHeaderThreshold) {
        grid-column: 4;
      }
      justify-content: right;

      .currency {
        @media (max-width: $mobileHeaderThreshold) {
          display: none;
        }
        color: var(--header-top-font-color);
      }
      .logged-out {
        white-space: nowrap;
        display: flex;
        align-items: center;
        gap: var(--padding-small);
        .login {
          font-weight: var(--main_header_bold_font_weight);
        }
        .register {
          padding: 8px;
        }
      }
      .user-menu {
        @media (max-width: $smallHeaderThreshold) {
          display: none;
        }
        .menu-component {
          padding-right: 0px;
          color: var(--header-top-font-color);
          .menu-icon {
            color: var(--header-top-font-color);
            width: 19px;
            height: 18px;
          }
        }
      }
      .burger {
        @media (min-width: $smallHeaderThreshold) {
          display: none;
        }
        height: 24px;
        cursor: pointer;
      }
    }

    .logo {
      grid-column: 2;
      @media (min-width: $smallHeaderThreshold) {
        @include bottom-cell(2);
      }
      @media (max-width: $smallHeaderThreshold) {
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
      position: relative;
      @media (min-width: $smallHeaderThreshold) {
        grid-column: 3;
        grid-column-end: span 3;
        @include bottom-cell(2);
        justify-content: flex-end;
        .decoration {
          display: none;
        }
      }
      @media (max-width: $smallHeaderThreshold) {
        grid-row: 2;
        grid-column: 1;
        grid-column-end: span 5;
        .decoration {
          position: absolute;
          top: 0px;
          bottom: -1px;
          left: calc(1px - var(--content-margin));
          right: calc(1px - var(--content-margin));
          border-bottom-left-radius: var(--border-radius);
          border-bottom-right-radius: var(--border-radius);
          border: 1px solid var(--primary-color);
          border-top: none;
        }
      }
    }
  }
}
</style>
