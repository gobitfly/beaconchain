<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faDiscord, faTwitter, faGithub } from '@fortawesome/free-brands-svg-icons'
import {
  faCaretRight,
  faHistory,
  faCube,
  faCubes,
  faCreditCard,
  faUpload,
  faTable,
  faUserSlash,
  faMedal,
  faFileImport,
  faFileSignature,
  faMoneyBill,
  faChartBar,
  faChartLine,
  faFireFlame,
  faRobot,
  faMegaphone,
  faExternalLinkAlt,
  faDesktop
} from '@fortawesome/pro-solid-svg-icons'
import {
  faDrumstickBite,
  faChartPie,
  faRocket,
  faMoneyBill as farMoneyBill,
  faCalculator,
  faProjectDiagram,
  faMobileScreen,
  faGem,
  faLaptopCode,
  faGasPump,
  faBell,
  faPaintBrush,
  faDesktop as farDesktop
} from '@fortawesome/pro-regular-svg-icons'
import {
  faBuildingColumns
} from '@fortawesome/sharp-solid-svg-icons'

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
import { mobileHeaderThreshold, smallHeaderThreshold } from '~/types/header'

const { t: $t } = useI18n()
const { width } = useWindowSize()
const { doLogout, isLoggedIn } = useUserStore()
const { withLabel, currency, setCurrency } = useCurrency()
const route = useRoute()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const megaMenu = ref<{toggle:(evt:Event)=>void, mobileActive: boolean} | null>(null)

const breakpoint = `${smallHeaderThreshold}px`
const isSmallScreen = computed(() => width.value < smallHeaderThreshold)
const isMobile = computed(() => width.value < mobileHeaderThreshold)

const items = computed(() => {
  let list: MenuItem[] = []

  if (showInDevelopment) {
    list = [
      {
        label: 'Ethereum',
        items: [
          [
            {
              label: $t('header.megamenu.blockchain'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  svg: NetworkEthereum,
                  url: '/'
                },
                {
                  label: $t('header.megamenu.epochs'),
                  icon: faHistory,
                  url: '/epochs'
                },
                {
                  label: $t('header.megamenu.slots'),
                  icon: faCube,
                  url: '/slots'
                },
                {
                  label: $t('header.megamenu.blocks'),
                  icon: faCubes,
                  url: '/blocks'
                },
                {
                  label: $t('header.megamenu.txs'),
                  icon: faCreditCard,
                  url: '/transactions'
                },
                {
                  label: $t('header.megamenu.mempool'),
                  icon: faUpload,
                  url: '/mempool'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.validators'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  icon: faTable,
                  url: '/validators'
                },
                {
                  label: $t('header.megamenu.slashings'),
                  icon: faUserSlash,
                  url: '/validators/slashings'
                },
                {
                  label: $t('header.megamenu.validator_leaderboard'),
                  icon: faMedal,
                  url: '/validators/leaderboard'
                },
                {
                  label: $t('header.megamenu.deposit_leaderboard'),
                  icon: faFileImport,
                  url: '/validators/deposit-leaderboard'
                },
                {
                  label: $t('header.megamenu.deposits'),
                  icon: faFileSignature,
                  url: '/validators/deposits'
                },
                {
                  label: $t('header.megamenu.withdrawals'),
                  icon: faMoneyBill,
                  url: '/validators/withdrawals'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.staking_pools'),
              items: [
                {
                  label: $t('header.megamenu.run_a_validator'),
                  svg: IconEthermineStaking,
                  url: 'https://ethpool.org/',
                  class: 'orange-box',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.eth_store'),
                  svg: IconEthStore,
                  url: '/ethstore'
                },
                {
                  label: $t('header.megamenu.staking_services'),
                  icon: faDrumstickBite,
                  url: '/stakingServices'
                },
                {
                  label: $t('header.megamenu.pool_benchmarks'),
                  icon: faChartPie,
                  url: '/pools'
                },
                {
                  label: $t('header.megamenu.rocket_pool_stats'),
                  icon: faRocket,
                  url: '/pools/rocketpool'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.stats'),
              items: [
                {
                  label: $t('header.megamenu.charts'),
                  icon: faChartBar,
                  url: '/charts'
                },
                {
                  label: $t('header.megamenu.reward_history'),
                  icon: farMoneyBill,
                  url: '/rewards'
                },
                {
                  label: $t('header.megamenu.profit_calculator'),
                  icon: faCalculator,
                  url: '/calculator'
                },
                {
                  label: $t('header.megamenu.block_viz'),
                  icon: faProjectDiagram,
                  url: '/vis'
                },
                {
                  label: $t('header.megamenu.correlations'),
                  icon: faChartLine,
                  url: '/correlations'
                },
                {
                  label: $t('header.megamenu.eip1599_burn'),
                  icon: faFireFlame,
                  url: '/burn'
                },
                {
                  label: $t('header.megamenu.relays'),
                  icon: faRobot,
                  url: '/relays'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.tools'),
              items: [
                {
                  label: $t('header.megamenu.beaconchain_app'),
                  icon: faMobileScreen,
                  url: '/mobile'
                },
                {
                  label: $t('header.megamenu.beaconchain_premium'),
                  icon: faGem,
                  url: '/pricing'
                },
                {
                  label: $t('header.megamenu.webhooks'),
                  svg: IconWebhook,
                  url: '/user/webhooks'
                },
                {
                  label: $t('header.megamenu.api_docs'),
                  icon: faLaptopCode,
                  url: '/api/v1/docs/index.html'
                },
                {
                  // TODO: Requires pricing page to set the toggle at the top to "API Pricing"
                  label: $t('header.megamenu.api_pricing'),
                  icon: faLaptopCode,
                  url: '/pricing'
                },
                {
                  label: $t('header.megamenu.unit_converter'),
                  icon: faBuildingColumns,
                  url: '/tools/unitConverter'
                },
                {
                  label: $t('header.megamenu.gasnow'),
                  icon: faGasPump,
                  url: '/gasnow'
                },
                {
                  label: $t('header.megamenu.broadcast_signed_messages'),
                  icon: faMegaphone,
                  url: '/tools/broadcast'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.services'),
              items: [
                {
                  label: $t('header.megamenu.eversteel'),
                  svg: IconEversteel,
                  url: 'https://eversteel.io/',
                  class: 'orange-box',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.notifications'),
                  icon: faBell,
                  url: '/notifications'
                },
                {
                  label: $t('header.megamenu.graffiti_wall'),
                  icon: faPaintBrush,
                  url: '/graffitiwall'
                },
                {
                  label: $t('header.megamenu.ethereum_clients'),
                  icon: farDesktop,
                  url: '/ethClients'
                },
                {
                  label: $t('header.megamenu.knowledge_base'),
                  icon: faExternalLinkAlt,
                  url: 'https://kb.beaconcha.in',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.slot_finder'),
                  icon: faCube,
                  url: '/slots/finder'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.community'),
              items: [
                {
                  label: 'Discord',
                  icon: faDiscord,
                  url: 'https://dsc.gg/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Twitter',
                  icon: faTwitter,
                  url: 'https://twitter.com/beaconcha_in',
                  target: Target.External
                },
                {
                  label: 'Github',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Github Mobile App',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
                  target: Target.External
                }
              ]
            }
          ]]
      },
      {
        label: 'Gnosis',
        items: [
          [
            {
              label: $t('header.megamenu.blockchain'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  svg: NetworkGnosis,
                  url: '/'
                },
                {
                  label: $t('header.megamenu.epochs'),
                  icon: faHistory,
                  url: '/epochs'
                },
                {
                  label: $t('header.megamenu.slots'),
                  icon: faCube,
                  url: '/slots'
                },
                {
                  label: $t('header.megamenu.blocks'),
                  icon: faCubes,
                  url: '/blocks'
                },
                {
                  label: $t('header.megamenu.txs'),
                  icon: faCreditCard,
                  url: '/transactions'
                },
                {
                  label: $t('header.megamenu.validators'),
                  icon: faDesktop,
                  url: ''
                },
                {
                  label: $t('header.megamenu.mempool'),
                  icon: faUpload,
                  url: '/mempool'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.validators'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  icon: faTable,
                  url: '/validators'
                },
                {
                  label: $t('header.megamenu.slashings'),
                  icon: faUserSlash,
                  url: '/validators/slashings'
                },
                {
                  label: $t('header.megamenu.validator_leaderboard'),
                  icon: faMedal,
                  url: '/validators/leaderboard'
                },
                {
                  label: $t('header.megamenu.deposit_leaderboard'),
                  icon: faFileImport,
                  url: '/validators/deposit-leaderboard'
                },
                {
                  label: $t('header.megamenu.deposits'),
                  icon: faFileSignature,
                  url: '/validators/deposits'
                },
                {
                  label: $t('header.megamenu.withdrawals'),
                  icon: faMoneyBill,
                  url: '/validators/withdrawals'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.stats'),
              items: [
                {
                  label: $t('header.megamenu.charts'),
                  icon: faChartBar,
                  url: '/charts'
                },
                {
                  label: $t('header.megamenu.reward_history'),
                  icon: farMoneyBill,
                  url: '/rewards'
                },
                {
                  label: $t('header.megamenu.profit_calculator'),
                  icon: faCalculator,
                  url: '/calculator'
                },
                {
                  label: $t('header.megamenu.block_viz'),
                  icon: faProjectDiagram,
                  url: '/vis'
                },
                {
                  label: $t('header.megamenu.correlations'),
                  icon: faChartLine,
                  url: '/correlations'
                },
                {
                  label: $t('header.megamenu.eip1599_burn'),
                  icon: faFireFlame,
                  url: '/burn'
                },
                {
                  label: $t('header.megamenu.relays'),
                  icon: faRobot,
                  url: '/relays'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.tools'),
              items: [
                {
                  label: $t('header.megamenu.beaconchain_app'),
                  icon: faMobileScreen,
                  url: '/mobile'
                },
                {
                  label: $t('header.megamenu.beaconchain_premium'),
                  icon: faGem,
                  url: '/premium'
                },
                {
                  label: $t('header.megamenu.webhooks'),
                  svg: IconWebhook,
                  url: '/user/webhooks'
                },
                {
                  label: $t('header.megamenu.api_docs'),
                  icon: faLaptopCode,
                  url: '/api/v1/docs/index.html'
                },
                {
                  label: $t('header.megamenu.api_pricing'),
                  icon: faLaptopCode,
                  url: '/pricing'
                },
                {
                  label: $t('header.megamenu.broadcast_signed_messages'),
                  icon: faMegaphone,
                  url: '/tools/broadcast'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.services'),
              items: [
                {
                  label: $t('header.megamenu.eversteel'),
                  svg: IconEversteel,
                  url: 'https://eversteel.io/',
                  class: 'orange-box',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.notifications'),
                  icon: faBell,
                  url: '/notifications'
                },
                {
                  label: $t('header.megamenu.knowledge_base'),
                  icon: faExternalLinkAlt,
                  url: 'https://kb.beaconcha.in',
                  target: Target.External
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.community'),
              items: [
                {
                  label: 'Discord',
                  icon: faDiscord,
                  url: 'https://dsc.gg/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Twitter',
                  icon: faTwitter,
                  url: 'https://twitter.com/beaconcha_in',
                  target: Target.External
                },
                {
                  label: 'Github',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Github Mobile App',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
                  target: Target.External
                }
              ]
            }
          ]]
      },
      {
        label: 'Arbitrum',
        items: [
          [
            {
              label: $t('header.megamenu.blockchain'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  svg: NetworkArbitrum,
                  url: '/'
                },
                {
                  label: $t('header.megamenu.blocks'),
                  icon: faCubes,
                  url: '/blocks'
                },
                {
                  label: $t('header.megamenu.txs'),
                  icon: faCreditCard,
                  url: '/transactions'
                },
                {
                  label: $t('header.megamenu.txsL1L2'),
                  icon: faCreditCard,
                  url: ''
                },
                {
                  label: $t('header.megamenu.txsL2L1'),
                  icon: faCreditCard,
                  url: ''
                },
                {
                  label: $t('header.megamenu.mempool'),
                  icon: faUpload,
                  url: '/mempool'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.tools'),
              items: [
                {
                  label: $t('header.megamenu.webhooks'),
                  svg: IconWebhook,
                  url: '/user/webhooks'
                },
                {
                  label: $t('header.megamenu.api_docs'),
                  icon: faLaptopCode,
                  url: '/api/v1/docs/index.html'
                },
                {
                  label: $t('header.megamenu.api_pricing'),
                  icon: faLaptopCode,
                  url: '/pricing'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.services'),
              items: [
                {
                  label: $t('header.megamenu.eversteel'),
                  svg: IconEversteel,
                  url: 'https://eversteel.io/',
                  class: 'orange-box',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.knowledge_base'),
                  icon: faExternalLinkAlt,
                  url: 'https://kb.beaconcha.in',
                  target: Target.External
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.community'),
              items: [
                {
                  label: 'Discord',
                  icon: faDiscord,
                  url: 'https://dsc.gg/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Twitter',
                  icon: faTwitter,
                  url: 'https://twitter.com/beaconcha_in',
                  target: Target.External
                },
                {
                  label: 'Github',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Github Mobile App',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
                  target: Target.External
                }
              ]
            }
          ]]
      },
      {
        label: 'Base',
        items: [
          [
            {
              label: $t('header.megamenu.blockchain'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  svg: NetworkBase,
                  url: '/'
                },
                {
                  label: $t('header.megamenu.blocks'),
                  icon: faCubes,
                  url: '/blocks'
                },
                {
                  label: $t('header.megamenu.txs'),
                  icon: faCreditCard,
                  url: '/transactions'
                },
                {
                  label: $t('header.megamenu.txsL1L2'),
                  icon: faCreditCard,
                  url: ''
                },
                {
                  label: $t('header.megamenu.txsL2L1'),
                  icon: faCreditCard,
                  url: ''
                },
                {
                  label: $t('header.megamenu.mempool'),
                  icon: faUpload,
                  url: '/mempool'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.tools'),
              items: [
                {
                  label: $t('header.megamenu.webhooks'),
                  svg: IconWebhook,
                  url: '/user/webhooks'
                },
                {
                  label: $t('header.megamenu.api_docs'),
                  icon: faLaptopCode,
                  url: '/api/v1/docs/index.html'
                },
                {
                  label: $t('header.megamenu.api_pricing'),
                  icon: faLaptopCode,
                  url: '/pricing'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.services'),
              items: [
                {
                  label: $t('header.megamenu.eversteel'),
                  svg: IconEversteel,
                  url: 'https://eversteel.io/',
                  class: 'orange-box',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.knowledge_base'),
                  icon: faExternalLinkAlt,
                  url: 'https://kb.beaconcha.in',
                  target: Target.External
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.community'),
              items: [
                {
                  label: 'Discord',
                  icon: faDiscord,
                  url: 'https://dsc.gg/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Twitter',
                  icon: faTwitter,
                  url: 'https://twitter.com/beaconcha_in',
                  target: Target.External
                },
                {
                  label: 'Github',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Github Mobile App',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
                  target: Target.External
                }
              ]
            }
          ]]
      },
      {
        label: 'Optimism',
        items: [
          [
            {
              label: $t('header.megamenu.blockchain'),
              items: [
                {
                  label: $t('header.megamenu.overview'),
                  svg: NetworkOptimism,
                  url: '/'
                },
                {
                  label: $t('header.megamenu.blocks'),
                  icon: faCubes,
                  url: '/blocks'
                },
                {
                  label: $t('header.megamenu.txs'),
                  icon: faCreditCard,
                  url: '/transactions'
                },
                {
                  label: $t('header.megamenu.txsL1L2'),
                  icon: faCreditCard,
                  url: ''
                },
                {
                  label: $t('header.megamenu.txsL2L1'),
                  icon: faCreditCard,
                  url: ''
                },
                {
                  label: $t('header.megamenu.mempool'),
                  icon: faUpload,
                  url: '/mempool'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.tools'),
              items: [
                {
                  label: $t('header.megamenu.webhooks'),
                  svg: IconWebhook,
                  url: '/user/webhooks'
                },
                {
                  label: $t('header.megamenu.api_docs'),
                  icon: faLaptopCode,
                  url: '/api/v1/docs/index.html'
                },
                {
                  label: $t('header.megamenu.api_pricing'),
                  icon: faLaptopCode,
                  url: '/pricing'
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.services'),
              items: [
                {
                  label: $t('header.megamenu.eversteel'),
                  svg: IconEversteel,
                  url: 'https://eversteel.io/',
                  class: 'orange-box',
                  target: Target.External
                },
                {
                  label: $t('header.megamenu.knowledge_base'),
                  icon: faExternalLinkAlt,
                  url: 'https://kb.beaconcha.in',
                  target: Target.External
                }
              ]
            }
          ],
          [
            {
              label: $t('header.megamenu.community'),
              items: [
                {
                  label: 'Discord',
                  icon: faDiscord,
                  url: 'https://dsc.gg/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Twitter',
                  icon: faTwitter,
                  url: 'https://twitter.com/beaconcha_in',
                  target: Target.External
                },
                {
                  label: 'Github',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/beaconchain',
                  target: Target.External
                },
                {
                  label: 'Github Mobile App',
                  icon: faGithub,
                  url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
                  target: Target.External
                }
              ]
            }
          ]
        ]
      },
      {
        label: $t('header.megamenu.dashboard'),
        url: '/dashboard'
      },
      {
        label: $t('header.megamenu.pricing'),
        url: '/pricing'
      },
      {
        label: $t('header.megamenu.notifications'),
        url: '/notifications'
      }
    ]
  } else {
    list = [
      {
        label: $t('header.megamenu.dashboard'),
        url: '/dashboard'
      },
      {
        label: $t('header.megamenu.pricing'),
        url: '/pricing'
      }
    ]
  }
  if (isMobile.value) {
    list.push({
      label: currency.value,
      currency: currency.value,
      items: [[{
        label: $t('header.megamenu.select_currency'),
        items: withLabel.value.map(m => ({ ...m, command: () => setCurrency(m.currency) }))
      }]]
    })
  }
  if (isSmallScreen.value && isLoggedIn.value) {
    list.push(
      {
        label: $t('header.logout'),
        command: () => doLogout()
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
  toggleMegaMenu,
  isMobileMenuOpen
})
</script>

<template>
  <ClientOnly>
    <MegaMenu ref="megaMenu" :model="items" :breakpoint="breakpoint">
      <template #item="{item, hasSubmenu}">
        <span class="p-menuitem-link">
          <span v-if="item.svg || item.icon || item.currency" class="p-menuitem-icon iconSpacing" data-pc-section="icon">
            <component :is="item.svg" v-if="item.svg" class="monochromatic" />
            <FontAwesomeIcon v-else-if="item.icon" class="icon" :icon="item.icon" />
            <IconCurrency v-else-if="item.currency" :currency="item.currency" />
          </span>
          <BcLink v-if="item.url" :to="item.url" :replace="route.path.startsWith(item.url)">
            <span :class="[item.class]" class="p-menuitem-text">
              <span>{{ item.label }}</span>
            </span>
          </BcLink>
          <div v-else class="pointer p-menuitem-text" :class="[item.class]" @click="item.command?.(null as any)">
            {{ item.label }}
          </div>
          <FontAwesomeIcon v-if="hasSubmenu" :icon="faCaretRight" class="p-icon p-submenu-icon" />
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
