<script setup lang="ts">
import type { BcHeaderMegaMenu } from '#build/components'
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { useNetworkStore } from '~/stores/useNetworkStore'
import {
  mobileHeaderThreshold, smallHeaderThreshold,
} from '~/types/header'

defineProps<{
  isHomePage: boolean,
  minimalist: boolean,
}>()
const latestStateStore = useLatestStateStore()
const { latestState } = storeToRefs(latestStateStore)
const {
  getEpochFromSlot,
  networkInfo,
} = useNetworkStore()
const {
  doLogout,
  isLoggedIn,
} = useUserStore()
const {
  displayCurrencyDefault,
  exchangeRates,
  formatAmount,
  selectedCurrencyMain,
} = useCurrency()
const { width } = useWindowSize()
const { t: $t } = useTranslation()

const isSmallScreen = computed(() => width.value < smallHeaderThreshold)
const isMobileScreen = computed(() => width.value < mobileHeaderThreshold)

const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const hideInDevelopmentClass = showInDevelopment
  ? ''
  : 'hide-because-it-is-unfinished' // TODO: once the searchbar is enabled in production, delete this line

const megaMenu = ref<null | typeof BcHeaderMegaMenu>(null)

const hasExchangeRates = computed(() => exchangeRates.value.length > 1)

const currentRate = computed(() => {
  if (selectedCurrencyMain.value === displayCurrencyDefault.main) {
    return formatAmount('1', {
      sourceCurrency: displayCurrencyDefault.main,
      sourceUnit: 'base',
      targetCurrency: displayCurrencyDefault.fiat,
    })
  }
  return formatAmount('1', {
    sourceCurrency: displayCurrencyDefault.main,
    sourceUnit: 'base',
  })
})

const currentEpoch = computed(() =>
  latestState.value?.current_slot !== undefined
    ? getEpochFromSlot(latestState.value.current_slot)
    : undefined,
)

const toggleMegaMenu = (evt: Event) => {
  megaMenu.value?.toggleMegaMenu(evt)
}

const isMobileMegaMenuOpen = computed(() => megaMenu.value?.isMobileMenuOpen)

const v1Domain = useV1Domain()

type UserMenuItem = { command: () => Promise<void>, label: string }
const userMenu: UserMenuItem[] = [
  {
    command: async () => { await navigateTo(`${v1Domain}/user/settings`, { external: true }) },
    label: $t('header.settings'),
  },
  {
    command: () => doLogout(),
    label: $t('header.logout'),
  },
]
const handleUserMenuSelect = async (value: UserMenuItem) => {
  await value.command?.()
}
const { url } = useV1Login()
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
        <span v-if="latestState?.current_slot"><span>{{ $t("header.slot") }}</span>:
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
        <span v-if="currentEpoch !== undefined"><span>{{ $t("header.epoch") }}</span>:
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
        <span
          v-if="hasExchangeRates"
          class="currency-info"
        >
          <BcCurrencyIcon
            :currency-code="displayCurrencyDefault.main"
            class="network-icon"
          />
          {{ displayCurrencyDefault.main }}:
          <span
            class="bold"
          >
            {{ currentRate }}
          </span>
        </span>
      </div>

      <div class="grid-cell controls">
        <BcCurrencySelection
          v-if="hasExchangeRates"
          class="currency"
          :show-currency-icon="!isMobileScreen"
        />
        <div
          v-if="!isLoggedIn"
          class="logged-out"
        >
          <BcLink
            :to="url"
          >
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
            panel-class="user-menu-panel"
            @select="handleUserMenuSelect"
          >
            <template #value>
              <BcIcon
                class="menu-icon"
                name="circle-user"
              />
            </template>
          </BcDropdown>
        </div>
        <BcButtonIcon
          class="burger"
          name="bars"
          screenreader-text="header.open_navigation"
          @click.stop="toggleMegaMenu"
        />
      </div>

      <div class="grid-cell explorer-info">
        <BcHeaderLogo layout-adaptability="high" />
        <span class="variant">
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
  &.hide-because-it-is-unfinished {
    // TODO: once the searchbar is enabled in production, delete this block (because border-bottom is always needed, due to the fact that the lower header is always visible (it contains the search bar when the screeen is narrow, otherwise the logo and mega menu))
    @media (max-width: $smallHeaderThreshold) {
      border-bottom: none;
    }
  }

  .rows {
    position: relative;
    display: grid;
    grid-template-columns: 0px min-content min-content auto min-content 0px; // the 0px are paddings, useless now but they exist in the structure of the grid so ready to be set if they are wanted one day
    grid-template-rows: var(--navbar-height) minmax(
        var(--navbar2-height),
        min-content
      );
    width: var(--content-width);
    color: var(--header-top-font-color);
    font-family: var(--main_header_font_family);
    font-size: var(--main_header_font_size);
    font-weight: var(--main_header_font_weight);
    color: var(--header-top-font-color);
    @media (max-width: $smallHeaderThreshold) {
      grid-template-columns: 0px min-content auto min-content 0px; // same remark about the 0px
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
      .currency-info {
        display: flex;
        align-items: center;
        gap: var(--padding-small);
      }
      .network-icon {
        width: 1.25rem;
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
      :global(.user-menu-panel) {
        // hack: panel should always get opened to the left, but this is not possible with component props
        $widthThePanelHasEnoughSpaceToOpenToTheRight: 1558px;
        $widthOfUserMenu: 43px;
        @media screen and (min-width: $widthThePanelHasEnoughSpaceToOpenToTheRight) {
          translate: calc(-100% + $widthOfUserMenu);
        }
      }
      .burger {
        color: var(--header-top-font-color);
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
        .large-screen {
          display: inline;
        }
        .mobile {
          display: none;
        }
        @media (max-width: $smallHeaderThreshold) {
          color: var(--grey);
        }
        @media (max-width: $mobileHeaderThreshold) {
          margin-bottom: auto;
          font-size: var(--button_font_size);
          .large-screen {
            display: none;
          }
          .mobile {
            display: inline;
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
