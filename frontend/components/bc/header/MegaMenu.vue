<script setup lang="ts">
import IconEthereumOverview from '~/components/icon/megaMenu/EthereumOverview.vue'
import IconGnosisOverview from '~/components/icon/megaMenu/GnosisOverview.vue'
import IconArbitrumOverview from '~/components/icon/megaMenu/ArbitrumOverview.vue'
import IconBaseOverview from '~/components/icon/megaMenu/BaseOverview.vue'
import IconOptimismOverview from '~/components/icon/megaMenu/OptimismOverview.vue'
import IconEthermineStaking from '~/components/icon/megaMenu/EthermineStaking.vue'
import IconEthStore from '~/components/icon/megaMenu/EthStore.vue'
import IconEversteel from '~/components/icon/megaMenu/EverSteel.vue'
import IconWebhook from '~/components/icon/megaMenu/WebHook.vue'

const { t: $t } = useI18n()

// const minFullViewWidth = '1360px'
const minFullViewWidth = '1360px'

const items = [
  {
    label: 'Ethereum',
    items: [
      [
        {
          label: $t('header.megamenu.blockchain'),
          items: [
            {
              label: $t('header.megamenu.overview'),
              svg: IconEthereumOverview,
              url: '/'
            },
            {
              label: $t('header.megamenu.epochs'),
              icon: 'fa-solid fa-history',
              url: '/epochs'
            },
            {
              label: $t('header.megamenu.slots'),
              icon: 'fas fa-cube',
              url: '/slots'
            },
            {
              label: $t('header.megamenu.blocks'),
              icon: 'fas fa-cubes',
              url: '/blocks'
            },
            {
              label: $t('header.megamenu.txs'),
              icon: 'fas fa-credit-card',
              url: '/transactions'
            },
            {
              label: $t('header.megamenu.mempool'),
              icon: 'fas fa-upload',
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
              icon: 'fas fa-table',
              url: '/validators'
            },
            {
              label: $t('header.megamenu.slashings'),
              icon: 'fas fa-user-slash',
              url: '/validators/slashings'
            },
            {
              label: $t('header.megamenu.validator_leaderboard'),
              icon: 'fas fa-medal',
              url: '/validators/leaderboard'
            },
            {
              label: $t('header.megamenu.deposit_leaderboard'),
              icon: 'fas fa-file-import',
              url: '/validators/deposit-leaderboard'
            },
            {
              label: $t('header.megamenu.deposits'),
              icon: 'fas fa-file-signature',
              url: '/validators/deposits'
            },
            {
              label: $t('header.megamenu.withdrawals'),
              icon: 'fas fa-money-bill',
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
              class: 'orange-box'
            },
            {
              label: $t('header.megamenu.eth_store'),
              svg: IconEthStore,
              url: '/ethstore'
            },
            {
              label: $t('header.megamenu.staking_services'),
              icon: 'far fa-drumstick-bite',
              url: '/stakingServices'
            },
            {
              label: $t('header.megamenu.pool_benchmarks'),
              icon: 'far fa-chart-pie',
              url: '/pools'
            },
            {
              label: $t('header.megamenu.rocket_pool_stats'),
              icon: 'far fa-rocket',
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
              icon: 'fas fa-chart-bar',
              url: '/charts'
            },
            {
              label: $t('header.megamenu.reward_history'),
              icon: 'far fa-money-bill',
              url: '/rewards'
            },
            {
              label: $t('header.megamenu.profit_calculator'),
              icon: 'far fa-calculator',
              url: '/calculator'
            },
            {
              label: $t('header.megamenu.block_viz'),
              icon: 'far fa-project-diagram',
              url: '/vis'
            },
            {
              label: $t('header.megamenu.correlations'),
              icon: 'fas fa-chart-line',
              url: '/correlations'
            },
            {
              label: $t('header.megamenu.eip1599_burn'),
              icon: 'fas fa-fire-flame',
              url: '/burn'
            },
            {
              label: $t('header.megamenu.relays'),
              icon: 'fas fa-robot',
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
              icon: 'far fa-mobile-screen',
              url: '/mobile'
            },
            {
              label: $t('header.megamenu.beaconchain_premium'),
              icon: 'far fa-gem',
              url: '/premium'
            },
            {
              label: $t('header.megamenu.webhooks'),
              svg: IconWebhook,
              url: '/user/webhooks'
            },
            {
              label: $t('header.megamenu.api_docs'),
              icon: 'far fa-laptop-code',
              url: '/api/v1/docs/index.html'
            },
            {
              label: $t('header.megamenu.api_pricing'),
              icon: 'far fa-laptop-code',
              url: '/pricing'
            },
            {
              label: $t('header.megamenu.unit_converter'),
              icon: 'fass fa-building-columns',
              url: '/tools/unitConverter'
            },
            {
              label: $t('header.megamenu.gasnow'),
              icon: 'far fa-gas-pump',
              url: '/gasnow'
            },
            {
              label: $t('header.megamenu.broadcast_signed_messages'),
              icon: 'fas fa-megaphone',
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
              class: 'orange-box'
            },
            {
              label: $t('header.megamenu.notifications'),
              icon: 'far fa-bell',
              url: '/user/notifications'
            },
            {
              label: $t('header.megamenu.graffiti_wall'),
              icon: 'far fa-paint-brush',
              url: '/graffitiwall'
            },
            {
              label: $t('header.megamenu.ethereum_clients'),
              icon: 'far fa-desktop',
              url: '/ethClients'
            },
            {
              label: $t('header.megamenu.knowledge_base'),
              icon: 'fas fa-external-link-alt',
              url: 'https://kb.beaconcha.in'
            },
            {
              label: $t('header.megamenu.slot_finder'),
              icon: 'fas fa-cube',
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
              icon: 'fab fa-discord',
              url: 'https://dsc.gg/beaconchain'
            },
            {
              label: 'Twitter',
              icon: 'fab fa-twitter',
              url: 'https://twitter.com/beaconcha_in'
            },
            {
              label: 'Github',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/beaconchain'
            },
            {
              label: 'Github Mobile App',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app'
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
              svg: IconGnosisOverview,
              url: '/'
            },
            {
              label: $t('header.megamenu.epochs'),
              icon: 'fas fa-history',
              url: '/epochs'
            },
            {
              label: $t('header.megamenu.slots'),
              icon: 'fas fa-cube',
              url: '/slots'
            },
            {
              label: $t('header.megamenu.blocks'),
              icon: 'fas fa-cubes',
              url: '/blocks'
            },
            {
              label: $t('header.megamenu.txs'),
              icon: 'fas fa-credit-card',
              url: '/transactions'
            },
            {
              label: $t('header.megamenu.validators'),
              icon: 'fas fa-desktop',
              url: ''
            },
            {
              label: $t('header.megamenu.mempool'),
              icon: 'fas fa-upload',
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
              icon: 'fas fa-table',
              url: '/validators'
            },
            {
              label: $t('header.megamenu.slashings'),
              icon: 'fas fa-user-slash',
              url: '/validators/slashings'
            },
            {
              label: $t('header.megamenu.validator_leaderboard'),
              icon: 'fas fa-medal',
              url: '/validators/leaderboard'
            },
            {
              label: $t('header.megamenu.deposit_leaderboard'),
              icon: 'fas fa-file-import',
              url: '/validators/deposit-leaderboard'
            },
            {
              label: $t('header.megamenu.deposits'),
              icon: 'fas fa-file-signature',
              url: '/validators/deposits'
            },
            {
              label: $t('header.megamenu.withdrawals'),
              icon: 'fas fa-money-bill',
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
              icon: 'fas fa-chart-bar',
              url: '/charts'
            },
            {
              label: $t('header.megamenu.reward_history'),
              icon: 'far fa-money-bill',
              url: '/rewards'
            },
            {
              label: $t('header.megamenu.profit_calculator'),
              icon: 'far fa-calculator',
              url: '/calculator'
            },
            {
              label: $t('header.megamenu.block_viz'),
              icon: 'far fa-project-diagram',
              url: '/vis'
            },
            {
              label: $t('header.megamenu.correlations'),
              icon: 'fas fa-chart-line',
              url: '/correlations'
            },
            {
              label: $t('header.megamenu.eip1599_burn'),
              icon: 'fas fa-fire-flame',
              url: '/burn'
            },
            {
              label: $t('header.megamenu.relays'),
              icon: 'fas fa-robot',
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
              icon: 'far fa-mobile-screen',
              url: '/mobile'
            },
            {
              label: $t('header.megamenu.beaconchain_premium'),
              icon: 'far fa-gem',
              url: '/premium'
            },
            {
              label: $t('header.megamenu.webhooks'),
              svg: IconWebhook,
              url: '/user/webhooks'
            },
            {
              label: $t('header.megamenu.api_docs'),
              icon: 'far fa-laptop-code',
              url: '/api/v1/docs/index.html'
            },
            {
              label: $t('header.megamenu.api_pricing'),
              icon: 'far fa-laptop-code',
              url: '/pricing'
            },
            {
              label: $t('header.megamenu.broadcast_signed_messages'),
              icon: 'fas fa-megaphone',
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
              class: 'orange-box'
            },
            {
              label: $t('header.megamenu.notifications'),
              icon: 'far fa-bell',
              url: '/user/notifications'
            },
            {
              label: $t('header.megamenu.knowledge_base'),
              icon: 'fas fa-external-link-alt',
              url: 'https://kb.beaconcha.in'
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
              icon: 'fab fa-discord',
              url: 'https://dsc.gg/beaconchain'
            },
            {
              label: 'Twitter',
              icon: 'fab fa-twitter',
              url: 'https://twitter.com/beaconcha_in'
            },
            {
              label: 'Github',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/beaconchain'
            },
            {
              label: 'Github Mobile App',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app'
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
              svg: IconArbitrumOverview,
              url: '/'
            },
            {
              label: $t('header.megamenu.blocks'),
              icon: 'fas fa-cubes',
              url: '/blocks'
            },
            {
              label: $t('header.megamenu.txs'),
              icon: 'fas fa-credit-card',
              url: '/transactions'
            },
            {
              label: $t('header.megamenu.txsL1L2'),
              icon: 'fas fa-credit-card',
              url: ''
            },
            {
              label: $t('header.megamenu.txsL2L1'),
              icon: 'fas fa-credit-card',
              url: ''
            },
            {
              label: $t('header.megamenu.mempool'),
              icon: 'fas fa-upload',
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
              icon: 'far fa-laptop-code',
              url: '/api/v1/docs/index.html'
            },
            {
              label: $t('header.megamenu.api_pricing'),
              icon: 'far fa-laptop-code',
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
              class: 'orange-box'
            },
            {
              label: $t('header.megamenu.knowledge_base'),
              icon: 'fas fa-external-link-alt',
              url: 'https://kb.beaconcha.in'
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
              icon: 'fab fa-discord',
              url: 'https://dsc.gg/beaconchain'
            },
            {
              label: 'Twitter',
              icon: 'fab fa-twitter',
              url: 'https://twitter.com/beaconcha_in'
            },
            {
              label: 'Github',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/beaconchain'
            },
            {
              label: 'Github Mobile App',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app'
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
              svg: IconBaseOverview,
              url: '/'
            },
            {
              label: $t('header.megamenu.blocks'),
              icon: 'fas fa-cubes',
              url: '/blocks'
            },
            {
              label: $t('header.megamenu.txs'),
              icon: 'fas fa-credit-card',
              url: '/transactions'
            },
            {
              label: $t('header.megamenu.txsL1L2'),
              icon: 'fas fa-credit-card',
              url: ''
            },
            {
              label: $t('header.megamenu.txsL2L1'),
              icon: 'fas fa-credit-card',
              url: ''
            },
            {
              label: $t('header.megamenu.mempool'),
              icon: 'fas fa-upload',
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
              icon: 'far fa-laptop-code',
              url: '/api/v1/docs/index.html'
            },
            {
              label: $t('header.megamenu.api_pricing'),
              icon: 'far fa-laptop-code',
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
              class: 'orange-box'
            },
            {
              label: $t('header.megamenu.knowledge_base'),
              icon: 'fas fa-external-link-alt',
              url: 'https://kb.beaconcha.in'
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
              icon: 'fab fa-discord',
              url: 'https://dsc.gg/beaconchain'
            },
            {
              label: 'Twitter',
              icon: 'fab fa-twitter',
              url: 'https://twitter.com/beaconcha_in'
            },
            {
              label: 'Github',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/beaconchain'
            },
            {
              label: 'Github Mobile App',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app'
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
              svg: IconOptimismOverview,
              url: '/'
            },
            {
              label: $t('header.megamenu.blocks'),
              icon: 'fas fa-cubes',
              url: '/blocks'
            },
            {
              label: $t('header.megamenu.txs'),
              icon: 'fas fa-credit-card',
              url: '/transactions'
            },
            {
              label: $t('header.megamenu.txsL1L2'),
              icon: 'fas fa-credit-card',
              url: ''
            },
            {
              label: $t('header.megamenu.txsL2L1'),
              icon: 'fas fa-credit-card',
              url: ''
            },
            {
              label: $t('header.megamenu.mempool'),
              icon: 'fas fa-upload',
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
              icon: 'far fa-laptop-code',
              url: '/api/v1/docs/index.html'
            },
            {
              label: $t('header.megamenu.api_pricing'),
              icon: 'far fa-laptop-code',
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
              class: 'orange-box'
            },
            {
              label: $t('header.megamenu.knowledge_base'),
              icon: 'fas fa-external-link-alt',
              url: 'https://kb.beaconcha.in'
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
              icon: 'fab fa-discord',
              url: 'https://dsc.gg/beaconchain'
            },
            {
              label: 'Twitter',
              icon: 'fab fa-twitter',
              url: 'https://twitter.com/beaconcha_in'
            },
            {
              label: 'Github',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/beaconchain'
            },
            {
              label: 'Github Mobile App',
              icon: 'fab fa-github',
              url: 'https://github.com/gobitfly/eth2-beaconchain-explorer-app'
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
    label: $t('header.megamenu.notifications'),
    url: '/user/notifications'
  }
]
</script>

<template>
  <MegaMenu :model="items" :breakpoint="minFullViewWidth">
    <template #itemicon="{ item }">
      <span v-if="item.svg || item.icon" class="p-menuitem-icon iconSpacing" data-pc-section="icon">
        <component :is="item.svg" v-if="item.svg" />
        <i v-else-if="item.icon" :class="item.icon" />
      </span>
    </template>
  </MegaMenu>
</template>

<style lang="scss"  scoped>
.iconSpacing {
  width: 25px;
  position: relative;

  svg,
  i {
    position: absolute;
    transform: translateY(-50%);
  }
}
</style>
