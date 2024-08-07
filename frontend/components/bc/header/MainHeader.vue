<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faBars,
  faCircleUser,
} from '@fortawesome/pro-solid-svg-icons'
import type { BcHeaderMegaMenu } from '#build/components'
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { useNetworkStore } from '~/stores/useNetworkStore'
import { SearchbarShape, SearchbarColors } from '~/types/searchbar'
import { mobileHeaderThreshold, smallHeaderThreshold } from '~/types/header'

defineProps<{
  isHomePage: boolean
  minimalist: boolean
}>()
const { latestState } = useLatestStateStore()
const { slotToEpoch, currentNetwork, networkInfo } = useNetworkStore()
const { doLogout, isLoggedIn } = useUserStore()
const { currency, available, rates } = useCurrency()
const { width } = useWindowSize()
const { t: $t } = useTranslation()

const colorMode = useColorMode()
const isSmallScreen = computed(() => width.value < smallHeaderThreshold)
const isMobileScreen = computed(() => width.value < mobileHeaderThreshold)

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const hideInDevelopmentClass = showInDevelopment ? '' : 'hide-because-it-is-unfinished' // TODO: once the searchbar is enabled in production, delete this line

const megaMenu = ref<typeof BcHeaderMegaMenu | null>(null)

const rate = computed(() => {
  if (isFiat(currency.value) && rates.value?.[currency.value]) {
    return rates.value[currency.value]
  }
  else if (rates.value?.USD) {
    return rates.value.USD
  }
  const fiat = available.value?.find(c => isFiat(c))
  if (fiat && rates.value?.[fiat]) {
    return rates.value[fiat]
  }
  return undefined
})

const currentEpoch = computed(() => latestState.value?.current_slot !== undefined ? slotToEpoch(latestState.value.current_slot) : undefined)

const toggleMegaMenu = (evt: Event) => {
  megaMenu.value?.toggleMegaMenu(evt)
}

const isMobileMegaMenuOpen = computed(() => megaMenu.value?.isMobileMenuOpen)

const userMenu = computed(() => {
  return [
    {
      label: $t('header.settings'),
      command: async () => { await navigateTo('../user/settings') },
    },
    {
      label: $t('header.logout'),
      command: () => doLogout(),
    },
  ]
})
</script>

<template>
  <div
    v-if="minimalist"
    class="minimalist"
  >
    <div class="top-background" />
    <div class="rows">
      <BcHeaderLogo layout-adaptability="low" />
    </div>
  </div>

  <div
    v-else
    class="complete"
    :class="hideInDevelopmentClass"
  >
    <div class="top-background" />
    <div class="rows">
      <div class="grid-cell blockchain-info">
        <span v-if="latestState?.current_slot"><span>{{ $t('header.current_slot') }}</span>:
          <BcLink
            :to="`/slot/${latestState.current_slot}`"
            :disabled="!showInDevelopment || null"
          >
            <BcFormatNumber
              class="bold"
              :value="latestState.current_slot"
            />
          </BcLink>
        </span>
        <span v-if="currentEpoch !== undefined"><span>{{ $t('header.current_epoch') }}</span>:
          <BcLink
            :to="`/epoch/${currentEpoch}`"
            :disabled="!showInDevelopment || null"
          >
            <BcFormatNumber
              class="bold"
              :value="currentEpoch"
            />
          </BcLink>
        </span>
        <span v-if="rate">
          <span>
            <IconNetwork
              :chain-id="currentNetwork"
              class="network-icon"
              :harmonize-perceived-size="true"
              :colored="false"
            />{{ networkInfo.elCurrency }}
          </span>:
          <span> {{ rate.symbol }}
            <BcFormatNumber
              class="bold"
              :value="rate.rate"
              :max-decimals="2"
            />
          </span>
        </span>
      </div>

      <div class="grid-cell search-bar">
        <BcSearchbarGeneral
          v-if="showInDevelopment && !isHomePage"
          class="bar"
          :bar-shape="SearchbarShape.Medium"
          :color-theme="isSmallScreen && colorMode.value != 'dark' ? SearchbarColors.LightBlue : SearchbarColors.DarkBlue"
          :screen-width-causing-sudden-change="smallHeaderThreshold"
        />
      </div>

      <div class="grid-cell controls">
        <BcCurrencySelection
          class="currency"
          :show-currency-icon="!isMobileScreen"
        />
        <div
          v-if="!isLoggedIn"
          class="logged-out"
        >
          <BcLink to="/login">
            <Button
              class="login"
              :label="$t('header.login')"
            />
          </BcLink>
        </div>
        <div
          v-else-if="!isSmallScreen"
          class="user-menu"
        >
          <BcDropdown
            :options="userMenu"
            variant="header"
            option-label="label"
            class="menu-component"
          >
            <template #value>
              <FontAwesomeIcon
                class="menu-icon"
                :icon="faCircleUser"
              />
            </template>
            <template #option="slotProps">
              <span @click="slotProps.command?.()">
                {{ slotProps.label }}
              </span>
            </template>
          </BcDropdown>
        </div>
        <FontAwesomeIcon
          :icon="faBars"
          class="burger"
          @click.stop.prevent="toggleMegaMenu"
        />
      </div>

      <div class="grid-cell explorer-info">
        <BcHeaderLogo layout-adaptability="high" />
        <span class="variant">
          v2 beta |
          <span class="mobile">{{ networkInfo.shortName }}</span>
          <span class="large-screen">{{ networkInfo.name }}</span>
        </span>
      </div>

      <div class="grid-cell mega-menu">
        <BcHeaderMegaMenu ref="megaMenu" />
        <div
          v-if="isMobileMegaMenuOpen"
          class="decoration"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

// do not change these two values without changing the values in HeaderLogo.vue and in types/header.ts accordingly
$mobileHeaderThreshold: 600px;
$smallHeaderThreshold: 1024px;

@mixin common {
  position: relative;
  display: flex;
  width: 100%;
  justify-content: center;
  .top-background {
    position: absolute;
    width: 100%;
    height: var(--navbar-height);
    background-color: var(--dark-blue);
  }
  .rows {
    width: var(--content-width);
  }
}

.minimalist {
  color: var(--header-top-font-color);
  @include common();
  @media (max-width: $mobileHeaderThreshold) {
    .top-background {
      height: 36px;
    }
  }
}

.complete {
  top: -1px; // needed for some reason to perfectly match Figma
  border-bottom: 1px solid var(--container-border-color);
  background-color: var(--container-background);
  @include common();
  &.hide-because-it-is-unfinished {  // TODO: once the searchbar is enabled in production, delete this block (because border-bottom is always needed, due to the fact that the lower header is always visible (it contains the search bar when the screeen is narrow, otherwise the logo and mega menu))
    @media (max-width: $smallHeaderThreshold) {
      border-bottom: none;
    }
  }

  .rows {
    position: relative;
    display: grid;
    grid-template-columns: 0px min-content min-content auto min-content 0px;  // the 0px are paddings, useless now but they exist in the structure of the grid so ready to be set if they are wanted one day
    grid-template-rows: var(--navbar-height) minmax(var(--navbar2-height), min-content);
    width: var(--content-width);
    color: var(--header-top-font-color);
    font-family: var(--main_header_font_family);
    font-size: var(--main_header_font_size);
    font-weight: var(--main_header_font_weight);
    color: var(--header-top-font-color);
    @media (max-width: $smallHeaderThreshold) {
      grid-template-columns: 0px min-content auto min-content 0px;  // same remark about the 0px
      grid-template-rows: var(--navbar-height) min-content;
    }
    @mixin bottom-cell($row) {
      color: var(--container-color);
      grid-row: $row;
    }
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
      margin-right: var(--padding-large);
      @media (min-width: $smallHeaderThreshold) {
        grid-row: 1;
        grid-column: 2;
        grid-column-end: span 2;
      }
      @media (max-width: $smallHeaderThreshold) {
        display: none;
      }
      .network-icon {
        vertical-align: middle;
        height: 18px;
        width: 18px;
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
        margin-top: var(--content-margin);
        margin-bottom: var(--content-margin);
        @media (min-width: $smallHeaderThreshold) {
          max-width: 460px;
        }
      }
    }

    .controls {
      user-select: none;
      grid-row: 1;
      grid-column: 5;
      justify-content: right;
      @media (max-width: $smallHeaderThreshold) {
        grid-column: 4;
      }

      .currency {
        color: var(--header-top-font-color);
      }
      .logged-out {
        white-space: nowrap;
        display: flex;
        align-items: center;
        gap: var(--padding-small);
        .login {
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
        height: 24px;
        cursor: pointer;
        @media (min-width: $smallHeaderThreshold) {
          display: none;
        }
      }
    }

    .explorer-info {
      grid-column: 2;
      height: unset;
      @media (min-width: $smallHeaderThreshold) {
        @include bottom-cell(2);
      }
      @media (max-width: $smallHeaderThreshold) {
        grid-row: 1;
      }
      .variant {
        position: relative;
        margin-top: auto;
        font-size: var(--tiny_text_font_size);
        color: var(--megamenu-text-color);
        line-height: 10px;
        .large-screen { display: inline }
        .mobile { display: none }
        @media (max-width: $smallHeaderThreshold) {
          color: var(--grey);
        }
        @media (max-width: $mobileHeaderThreshold) {
          margin-bottom: auto;
          font-size: var(--button_font_size);
          .large-screen { display: none }
          .mobile { display: inline }
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
