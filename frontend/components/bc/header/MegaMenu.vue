<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faDiscord,
  faGithub,
  faTwitter,
} from '@fortawesome/free-brands-svg-icons'
import {
  faCaretRight,
  faChartBar,
  faChartLine,
  faCreditCard,
  faCube,
  faCubes,
  faDesktop,
  faExternalLinkAlt,
  faFileImport,
  faFileSignature,
  faFireFlame,
  faHistory,
  faMedal,
  faMegaphone,
  faMoneyBill,
  faRobot,
  faTable,
  faUpload,
  faUserSlash,
} from '@fortawesome/pro-solid-svg-icons'
import {
  faBell,
  faCalculator,
  faChartPie,
  faDrumstickBite,
  faGasPump,
  faGem,
  faLaptopCode,
  faMobileScreen,
  faPaintBrush,
  faProjectDiagram,
  faDesktop as farDesktop,
  faMoneyBill as farMoneyBill,
  faRocket,
} from '@fortawesome/pro-regular-svg-icons'
import { faBuildingColumns } from '@fortawesome/sharp-solid-svg-icons'

import type { MenuItem } from 'primevue/menuitem'
import MegaMenu from 'primevue/megamenu'
import NetworkEthereum from '~/components/icon/network/NetworkEthereum.vue'
import NetworkGnosis from '~/components/icon/network/NetworkGnosis.vue'
import NetworkArbitrum from '~/components/icon/network/NetworkArbitrum.vue'
import NetworkBase from '~/components/icon/network/NetworkBase.vue'
import NetworkOptimism from '~/components/icon/network/NetworkOptimism.vue'
import IconEthermineStaking from '~/components/icon/megaMenu/EthermineStaking.vue'
import IconEthStore from '~/components/icon/megaMenu/EthStore.vue'
import IconEversteel from '~/components/icon/megaMenu/EverSteel.vue'
import IconWebhook from '~/components/icon/megaMenu/WebHook.vue'

import { Target } from '~/types/links'
import {
  mobileHeaderThreshold, smallHeaderThreshold,
} from '~/types/header'

const { t: $t } = useTranslation()
const { width } = useWindowSize()
const {
  doLogout, isLoggedIn,
} = useUserStore()
const route = useRoute()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const megaMenu = ref<{
  mobileActive: boolean
  toggle: (evt: Event) => void
} | null>(null)

const breakpoint = `${smallHeaderThreshold}px`
const isSmallScreen = computed(() => width.value < smallHeaderThreshold)
const isMobile = computed(() => width.value < mobileHeaderThreshold)

const items = computed(() => {
  let list: MenuItem[] = []

  if (showInDevelopment) {
    list = [
      {
        items: [
          [ {
            items: [
              {
                label: $t('header.megamenu.overview'),
                svg: NetworkEthereum,
                url: '/',
              },
              {
                icon: faHistory,
                label: $t('header.megamenu.epochs'),
                url: '/epochs',
              },
              {
                icon: faCube,
                label: $t('header.megamenu.slots'),
                url: '/slots',
              },
              {
                icon: faCubes,
                label: $t('header.megamenu.blocks'),
                url: '/blocks',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txs'),
                url: '/transactions',
              },
              {
                icon: faUpload,
                label: $t('header.megamenu.mempool'),
                url: '/mempool',
              },
            ],
            label: $t('header.megamenu.blockchain'),
          } ],
          [ {
            items: [
              {
                icon: faTable,
                label: $t('header.megamenu.overview'),
                url: '/validators',
              },
              {
                icon: faUserSlash,
                label: $t('header.megamenu.slashings'),
                url: '/validators/slashings',
              },
              {
                icon: faMedal,
                label: $t('header.megamenu.validator_leaderboard'),
                url: '/validators/leaderboard',
              },
              {
                icon: faFileImport,
                label: $t('header.megamenu.deposit_leaderboard'),
                url: '/validators/deposit-leaderboard',
              },
              {
                icon: faFileSignature,
                label: $t('header.megamenu.deposits'),
                url: '/validators/deposits',
              },
              {
                icon: faMoneyBill,
                label: $t('header.megamenu.withdrawals'),
                url: '/validators/withdrawals',
              },
            ],
            label: $t('header.megamenu.validators'),
          } ],
          [ {
            items: [
              {
                class: 'orange-box',
                label: $t('header.megamenu.run_a_validator'),
                svg: IconEthermineStaking,
                target: Target.External,
                url: 'https://ethpool.org/',
              },
              {
                label: $t('header.megamenu.eth_store'),
                svg: IconEthStore,
                url: '/ethstore',
              },
              {
                icon: faDrumstickBite,
                label: $t('header.megamenu.staking_services'),
                url: '/stakingServices',
              },
              {
                icon: faChartPie,
                label: $t('header.megamenu.pool_benchmarks'),
                url: '/pools',
              },
              {
                icon: faRocket,
                label: $t('header.megamenu.rocket_pool_stats'),
                url: '/pools/rocketpool',
              },
            ],
            label: $t('header.megamenu.staking_pools'),
          } ],
          [ {
            items: [
              {
                icon: faChartBar,
                label: $t('header.megamenu.charts'),
                url: '/charts',
              },
              {
                icon: farMoneyBill,
                label: $t('header.megamenu.reward_history'),
                url: '/rewards',
              },
              {
                icon: faCalculator,
                label: $t('header.megamenu.profit_calculator'),
                url: '/calculator',
              },
              {
                icon: faProjectDiagram,
                label: $t('header.megamenu.block_viz'),
                url: '/vis',
              },
              {
                icon: faChartLine,
                label: $t('header.megamenu.correlations'),
                url: '/correlations',
              },
              {
                icon: faFireFlame,
                label: $t('header.megamenu.eip1599_burn'),
                url: '/burn',
              },
              {
                icon: faRobot,
                label: $t('header.megamenu.relays'),
                url: '/relays',
              },
            ],
            label: $t('header.megamenu.stats'),
          } ],
          [ {
            items: [
              {
                icon: faMobileScreen,
                label: $t('header.megamenu.beaconchain_app'),
                url: '/mobile',
              },
              {
                icon: faGem,
                label: $t('header.megamenu.beaconchain_premium'),
                url: '/pricing',
              },
              {
                label: $t('header.megamenu.webhooks'),
                svg: IconWebhook,
                url: '/user/webhooks',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_docs'),
                url: '/api/v1/docs/index.html',
              },
              {
                icon: faLaptopCode,
                // TODO: Requires pricing page to set the toggle at the top to "API Pricing"
                label: $t('header.megamenu.api_pricing'),
                url: '/pricing',
              },
              {
                icon: faBuildingColumns,
                label: $t('header.megamenu.unit_converter'),
                url: '/tools/unitConverter',
              },
              {
                icon: faGasPump,
                label: $t('header.megamenu.gasnow'),
                url: '/gasnow',
              },
              {
                icon: faMegaphone,
                label: $t('header.megamenu.broadcast_signed_messages'),
                url: '/tools/broadcast',
              },
            ],
            label: $t('header.megamenu.tools'),
          } ],
          [ {
            items: [
              {
                class: 'orange-box',
                label: $t('header.megamenu.eversteel'),
                svg: IconEversteel,
                target: Target.External,
                url: 'https://eversteel.io/',
              },
              {
                icon: faBell,
                label: $t('header.megamenu.notifications'),
                url: '/notifications',
              },
              {
                icon: faPaintBrush,
                label: $t('header.megamenu.graffiti_wall'),
                url: '/graffitiwall',
              },
              {
                icon: farDesktop,
                label: $t('header.megamenu.ethereum_clients'),
                url: '/ethClients',
              },
              {
                icon: faExternalLinkAlt,
                label: $t('header.megamenu.knowledge_base'),
                target: Target.External,
                url: 'https://kb.beaconcha.in',
              },
              {
                icon: faCube,
                label: $t('header.megamenu.slot_finder'),
                url: '/slots/finder',
              },
            ],
            label: $t('header.megamenu.services'),
          } ],
          [ {
            items: [
              {
                icon: faDiscord,
                label: 'Discord',
                target: Target.External,
                url: 'https://dsc.gg/beaconchain',
              },
              {
                icon: faTwitter,
                label: 'Twitter',
                target: Target.External,
                url: 'https://twitter.com/beaconcha_in',
              },
              {
                icon: faGithub,
                label: 'Github',
                target: Target.External,
                url: 'https://github.com/gobitfly/beaconchain',
              },
              {
                icon: faGithub,
                label: 'Github Mobile App',
                target: Target.External,
                url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
              },
            ],
            label: $t('header.megamenu.community'),
          } ],
        ],
        label: 'Ethereum',
      },
      {
        items: [
          [ {
            items: [
              {
                label: $t('header.megamenu.overview'),
                svg: NetworkGnosis,
                url: '/',
              },
              {
                icon: faHistory,
                label: $t('header.megamenu.epochs'),
                url: '/epochs',
              },
              {
                icon: faCube,
                label: $t('header.megamenu.slots'),
                url: '/slots',
              },
              {
                icon: faCubes,
                label: $t('header.megamenu.blocks'),
                url: '/blocks',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txs'),
                url: '/transactions',
              },
              {
                icon: faDesktop,
                label: $t('header.megamenu.validators'),
                url: '',
              },
              {
                icon: faUpload,
                label: $t('header.megamenu.mempool'),
                url: '/mempool',
              },
            ],
            label: $t('header.megamenu.blockchain'),
          } ],
          [ {
            items: [
              {
                icon: faTable,
                label: $t('header.megamenu.overview'),
                url: '/validators',
              },
              {
                icon: faUserSlash,
                label: $t('header.megamenu.slashings'),
                url: '/validators/slashings',
              },
              {
                icon: faMedal,
                label: $t('header.megamenu.validator_leaderboard'),
                url: '/validators/leaderboard',
              },
              {
                icon: faFileImport,
                label: $t('header.megamenu.deposit_leaderboard'),
                url: '/validators/deposit-leaderboard',
              },
              {
                icon: faFileSignature,
                label: $t('header.megamenu.deposits'),
                url: '/validators/deposits',
              },
              {
                icon: faMoneyBill,
                label: $t('header.megamenu.withdrawals'),
                url: '/validators/withdrawals',
              },
            ],
            label: $t('header.megamenu.validators'),
          } ],
          [ {
            items: [
              {
                icon: faChartBar,
                label: $t('header.megamenu.charts'),
                url: '/charts',
              },
              {
                icon: farMoneyBill,
                label: $t('header.megamenu.reward_history'),
                url: '/rewards',
              },
              {
                icon: faCalculator,
                label: $t('header.megamenu.profit_calculator'),
                url: '/calculator',
              },
              {
                icon: faProjectDiagram,
                label: $t('header.megamenu.block_viz'),
                url: '/vis',
              },
              {
                icon: faChartLine,
                label: $t('header.megamenu.correlations'),
                url: '/correlations',
              },
              {
                icon: faFireFlame,
                label: $t('header.megamenu.eip1599_burn'),
                url: '/burn',
              },
              {
                icon: faRobot,
                label: $t('header.megamenu.relays'),
                url: '/relays',
              },
            ],
            label: $t('header.megamenu.stats'),
          } ],
          [ {
            items: [
              {
                icon: faMobileScreen,
                label: $t('header.megamenu.beaconchain_app'),
                url: '/mobile',
              },
              {
                icon: faGem,
                label: $t('header.megamenu.beaconchain_premium'),
                url: '/premium',
              },
              {
                label: $t('header.megamenu.webhooks'),
                svg: IconWebhook,
                url: '/user/webhooks',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_docs'),
                url: '/api/v1/docs/index.html',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_pricing'),
                url: '/pricing',
              },
              {
                icon: faMegaphone,
                label: $t('header.megamenu.broadcast_signed_messages'),
                url: '/tools/broadcast',
              },
            ],
            label: $t('header.megamenu.tools'),
          } ],
          [ {
            items: [
              {
                class: 'orange-box',
                label: $t('header.megamenu.eversteel'),
                svg: IconEversteel,
                target: Target.External,
                url: 'https://eversteel.io/',
              },
              {
                icon: faBell,
                label: $t('header.megamenu.notifications'),
                url: '/notifications',
              },
              {
                icon: faExternalLinkAlt,
                label: $t('header.megamenu.knowledge_base'),
                target: Target.External,
                url: 'https://kb.beaconcha.in',
              },
            ],
            label: $t('header.megamenu.services'),
          } ],
          [ {
            items: [
              {
                icon: faDiscord,
                label: 'Discord',
                target: Target.External,
                url: 'https://dsc.gg/beaconchain',
              },
              {
                icon: faTwitter,
                label: 'Twitter',
                target: Target.External,
                url: 'https://twitter.com/beaconcha_in',
              },
              {
                icon: faGithub,
                label: 'Github',
                target: Target.External,
                url: 'https://github.com/gobitfly/beaconchain',
              },
              {
                icon: faGithub,
                label: 'Github Mobile App',
                target: Target.External,
                url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
              },
            ],
            label: $t('header.megamenu.community'),
          } ],
        ],
        label: 'Gnosis',
      },
      {
        items: [
          [ {
            items: [
              {
                label: $t('header.megamenu.overview'),
                svg: NetworkArbitrum,
                url: '/',
              },
              {
                icon: faCubes,
                label: $t('header.megamenu.blocks'),
                url: '/blocks',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txs'),
                url: '/transactions',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txsL1L2'),
                url: '',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txsL2L1'),
                url: '',
              },
              {
                icon: faUpload,
                label: $t('header.megamenu.mempool'),
                url: '/mempool',
              },
            ],
            label: $t('header.megamenu.blockchain'),
          } ],
          [ {
            items: [
              {
                label: $t('header.megamenu.webhooks'),
                svg: IconWebhook,
                url: '/user/webhooks',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_docs'),
                url: '/api/v1/docs/index.html',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_pricing'),
                url: '/pricing',
              },
            ],
            label: $t('header.megamenu.tools'),
          } ],
          [ {
            items: [
              {
                class: 'orange-box',
                label: $t('header.megamenu.eversteel'),
                svg: IconEversteel,
                target: Target.External,
                url: 'https://eversteel.io/',
              },
              {
                icon: faExternalLinkAlt,
                label: $t('header.megamenu.knowledge_base'),
                target: Target.External,
                url: 'https://kb.beaconcha.in',
              },
            ],
            label: $t('header.megamenu.services'),
          } ],
          [ {
            items: [
              {
                icon: faDiscord,
                label: 'Discord',
                target: Target.External,
                url: 'https://dsc.gg/beaconchain',
              },
              {
                icon: faTwitter,
                label: 'Twitter',
                target: Target.External,
                url: 'https://twitter.com/beaconcha_in',
              },
              {
                icon: faGithub,
                label: 'Github',
                target: Target.External,
                url: 'https://github.com/gobitfly/beaconchain',
              },
              {
                icon: faGithub,
                label: 'Github Mobile App',
                target: Target.External,
                url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
              },
            ],
            label: $t('header.megamenu.community'),
          } ],
        ],
        label: 'Arbitrum',
      },
      {
        items: [
          [ {
            items: [
              {
                label: $t('header.megamenu.overview'),
                svg: NetworkBase,
                url: '/',
              },
              {
                icon: faCubes,
                label: $t('header.megamenu.blocks'),
                url: '/blocks',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txs'),
                url: '/transactions',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txsL1L2'),
                url: '',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txsL2L1'),
                url: '',
              },
              {
                icon: faUpload,
                label: $t('header.megamenu.mempool'),
                url: '/mempool',
              },
            ],
            label: $t('header.megamenu.blockchain'),
          } ],
          [ {
            items: [
              {
                label: $t('header.megamenu.webhooks'),
                svg: IconWebhook,
                url: '/user/webhooks',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_docs'),
                url: '/api/v1/docs/index.html',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_pricing'),
                url: '/pricing',
              },
            ],
            label: $t('header.megamenu.tools'),
          } ],
          [ {
            items: [
              {
                class: 'orange-box',
                label: $t('header.megamenu.eversteel'),
                svg: IconEversteel,
                target: Target.External,
                url: 'https://eversteel.io/',
              },
              {
                icon: faExternalLinkAlt,
                label: $t('header.megamenu.knowledge_base'),
                target: Target.External,
                url: 'https://kb.beaconcha.in',
              },
            ],
            label: $t('header.megamenu.services'),
          } ],
          [ {
            items: [
              {
                icon: faDiscord,
                label: 'Discord',
                target: Target.External,
                url: 'https://dsc.gg/beaconchain',
              },
              {
                icon: faTwitter,
                label: 'Twitter',
                target: Target.External,
                url: 'https://twitter.com/beaconcha_in',
              },
              {
                icon: faGithub,
                label: 'Github',
                target: Target.External,
                url: 'https://github.com/gobitfly/beaconchain',
              },
              {
                icon: faGithub,
                label: 'Github Mobile App',
                target: Target.External,
                url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
              },
            ],
            label: $t('header.megamenu.community'),
          } ],
        ],
        label: 'Base',
      },
      {
        items: [
          [ {
            items: [
              {
                label: $t('header.megamenu.overview'),
                svg: NetworkOptimism,
                url: '/',
              },
              {
                icon: faCubes,
                label: $t('header.megamenu.blocks'),
                url: '/blocks',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txs'),
                url: '/transactions',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txsL1L2'),
                url: '',
              },
              {
                icon: faCreditCard,
                label: $t('header.megamenu.txsL2L1'),
                url: '',
              },
              {
                icon: faUpload,
                label: $t('header.megamenu.mempool'),
                url: '/mempool',
              },
            ],
            label: $t('header.megamenu.blockchain'),
          } ],
          [ {
            items: [
              {
                label: $t('header.megamenu.webhooks'),
                svg: IconWebhook,
                url: '/user/webhooks',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_docs'),
                url: '/api/v1/docs/index.html',
              },
              {
                icon: faLaptopCode,
                label: $t('header.megamenu.api_pricing'),
                url: '/pricing',
              },
            ],
            label: $t('header.megamenu.tools'),
          } ],
          [ {
            items: [
              {
                class: 'orange-box',
                label: $t('header.megamenu.eversteel'),
                svg: IconEversteel,
                target: Target.External,
                url: 'https://eversteel.io/',
              },
              {
                icon: faExternalLinkAlt,
                label: $t('header.megamenu.knowledge_base'),
                target: Target.External,
                url: 'https://kb.beaconcha.in',
              },
            ],
            label: $t('header.megamenu.services'),
          } ],
          [ {
            items: [
              {
                icon: faDiscord,
                label: 'Discord',
                target: Target.External,
                url: 'https://dsc.gg/beaconchain',
              },
              {
                icon: faTwitter,
                label: 'Twitter',
                target: Target.External,
                url: 'https://twitter.com/beaconcha_in',
              },
              {
                icon: faGithub,
                label: 'Github',
                target: Target.External,
                url: 'https://github.com/gobitfly/beaconchain',
              },
              {
                icon: faGithub,
                label: 'Github Mobile App',
                target: Target.External,
                url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
              },
            ],
            label: $t('header.megamenu.community'),
          } ],
        ],
        label: 'Optimism',
      },
      {
        label: $t('header.megamenu.dashboard'),
        url: '/dashboard',
      },
      {
        label: $t('header.megamenu.pricing'),
        url: '/pricing',
      },
      {
        label: $t('header.megamenu.notifications'),
        url: '/notifications',
      },
    ]
  }
  else {
    list = [
      {
        label: $t('header.megamenu.dashboard'),
        url: '/dashboard',
      },
      {
        label: $t('header.megamenu.pricing'),
        url: '/pricing',
      },
    ]
  }
  if (isMobile.value) {
    if (isLoggedIn.value) {
      list.push({
        command: async () => {
          await navigateTo('../user/settings')
        },
        label: $t('header.settings'),
      })
    }
  }
  if (isSmallScreen.value && isLoggedIn.value) {
    list.push({
      command: () => doLogout(),
      label: $t('header.logout'),
    })
  }
  return list
})

const isMobileMenuOpen = computed(() => megaMenu.value?.mobileActive)

const toggleMegaMenu = (evt: Event) => {
  megaMenu.value?.toggle(evt)
  document.body.focus()
}

defineExpose({
  isMobileMenuOpen,
  toggleMegaMenu,
})
</script>

<template>
  <ClientOnly>
    <MegaMenu
      ref="megaMenu"
      :model="items"
      :breakpoint="breakpoint"
    >
      <template #item="{ item, hasSubmenu }">
        <span class="p-menuitem-link">
          <span
            v-if="item.svg || item.icon"
            class="p-menuitem-icon iconSpacing"
            data-pc-section="icon"
          >
            <component
              :is="item.svg"
              v-if="item.svg"
              class="monochromatic"
            />
            <FontAwesomeIcon
              v-else-if="item.icon"
              class="icon"
              :icon="item.icon"
            />
          </span>
          <BcLink
            v-if="item.url"
            :to="item.url"
            :replace="route.path.startsWith(item.url)"
          >
            <span
              :class="[item.class]"
              class="p-menuitem-text"
            >
              <span>{{ item.label }}</span>
            </span>
          </BcLink>
          <div
            v-else
            class="pointer p-menuitem-text"
            :class="[item.class]"
            @click="item.command?.(null as any)"
          >
            {{ item.label }}
          </div>
          <FontAwesomeIcon
            v-if="hasSubmenu"
            :icon="faCaretRight"
            class="p-icon p-submenu-icon"
          />
        </span>
      </template>
    </MegaMenu>
  </ClientOnly>
</template>

<style lang="scss" scoped>
.iconSpacing {
  width: 25px;
  position: relative;

  img,
  svg,
  i {
    position: absolute;
    transform: translateY(-50%);
    max-width: 16px;
  }
}
</style>
