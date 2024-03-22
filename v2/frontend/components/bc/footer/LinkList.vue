<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faDiscord, faTwitter, faGithub } from '@fortawesome/free-brands-svg-icons'
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import {
  faBuilding,
  faFileContract,
  faUserSecret,
  faLaptopCode,
  faFileInvoiceDollar,
  faUserAstronaut,
  faAd,
  faShoppingCart,
  faCheckCircle
} from '@fortawesome/pro-solid-svg-icons'

import { Target } from '~/types/links'

const { t: $t } = useI18n()
type Row = { title: string, links: [string, IconDefinition, string, Target][] }
const columns: Row[] = [
  {
    title: $t('footer.legal_notices'),
    links: [
      [$t('footer.imprint'), faBuilding, '/imprint', Target.Internal],
      [$t('footer.terms'), faFileContract, 'https://storage.googleapis.com/legal.beaconcha.in/tos.pdf', Target.External],
      [$t('footer.privacy'), faUserSecret, 'https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf', Target.External]
    ]
  },
  {
    title: $t('footer.resources'),
    links: [
      [$t('footer.api_docs'), faLaptopCode, '/api/v2/docs/index.html', Target.Internal],
      [$t('footer.api_pricing'), faFileInvoiceDollar, '/pricing', Target.Internal],
      [$t('footer.premium'), faUserAstronaut, '/premium', Target.Internal],
      [$t('footer.advertise'), faAd, '/advertisewithus', Target.Internal],
      [$t('footer.shop'), faShoppingCart, 'https://shop.beaconcha.in', Target.External],
      [$t('footer.status'), faCheckCircle, 'https://status.beaconcha.in/', Target.External]
    ]
  },
  {
    title: $t('footer.links'),
    links: [
      ['Discord', faDiscord, 'https://dsc.gg/beaconchain', Target.External],
      ['Twitter', faTwitter, 'https://twitter.com/beaconcha_in', Target.External],
      ['Github', faGithub, 'https://github.com/gobitfly/beaconchain', Target.External],
      ['Github Mobile App', faGithub, 'https://github.com/gobitfly/eth2-beaconchain-explorer-app', Target.External]
      // [$t('footer.press_kit'), faNewspaper, '/presskit', Target.Internal] // TODO: Add link once press kit is available
    ]
  }
]
</script>

<template>
  <div v-for="column in columns" :key="column.title">
    <div class="title">
      {{ column.title }}
    </div>
    <div v-for="line in column.links" :key="line[0]" class="link-line">
      <NuxtLink :to="line[2]" :target="line[3]" class="link">
        <FontAwesomeIcon class="icon" :icon="line[1]" />
        {{ line[0] }}
      </NuxtLink>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.title {
  color: var(--Light-Grey);
  font-size: 20px;
  font-weight: bold;
  line-height: 33px;

  @media (min-width: 600px) {
    // large screen
    margin-bottom: 10px;
  }

  @media (max-width: 600px) {
    // mobile
    margin-top: 10px;
  }
}

.link-line {
  line-height: 27px;
  font-size: 16px;
}

.icon {
  display: inline-block;
  width: 20px;
  text-align: center;
}
</style>
