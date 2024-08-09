<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faDiscord,
  faGithub,
  faTwitter,
} from '@fortawesome/free-brands-svg-icons'
import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import {
  faBuilding,
  faCheckCircle,
  faFileContract,
  faShoppingCart,
  faUserAstronaut,
  faUserSecret,
} from '@fortawesome/pro-solid-svg-icons'

import { Target } from '~/types/links'

const { t: $t } = useTranslation()
type Row = { links: [string, IconDefinition, string, Target][],
  title: string, }
const columns: Row[] = [
  {
    links: [
      [
        $t('footer.imprint'),
        faBuilding,
        'https://beaconcha.in/imprint',
        Target.External,
      ],
      [
        $t('footer.terms'),
        faFileContract,
        'https://storage.googleapis.com/legal.beaconcha.in/tos.pdf',
        Target.External,
      ],
      [
        $t('footer.privacy'),
        faUserSecret,
        'https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf',
        Target.External,
      ],
    ],
    title: $t('footer.legal_notices'),
  },
  {
    links: [
      // TODO: Add link once API prices are available
      // [$t('footer.api_pricing'), faFileInvoiceDollar, '/pricing', Target.Internal],
      [
        $t('footer.premium'),
        faUserAstronaut,
        '/pricing',
        Target.Internal,
      ],
      // TODO: Add link once advertise page is available
      // [$t('footer.advertise'), faAd, '/advertisewithus', Target.Internal],
      [
        $t('footer.shop'),
        faShoppingCart,
        'https://shop.beaconcha.in',
        Target.External,
      ],
      [
        $t('footer.status'),
        faCheckCircle,
        'https://status.beaconcha.in/',
        Target.External,
      ],
    ],
    title: $t('footer.resources'),
  },
  {
    links: [
      [
        'Discord',
        faDiscord,
        'https://dsc.gg/beaconchain',
        Target.External,
      ],
      [
        'Twitter',
        faTwitter,
        'https://twitter.com/beaconcha_in',
        Target.External,
      ],
      [
        'Github',
        faGithub,
        'https://github.com/gobitfly/beaconchain',
        Target.External,
      ],
      [
        'Github Mobile App',
        faGithub,
        'https://github.com/gobitfly/eth2-beaconchain-explorer-app',
        Target.External,
      ],
      // TODO: Add link once press kit is available
      // [$t('footer.press_kit'), faNewspaper, '/presskit', Target.Internal]
    ],
    title: $t('footer.links'),
  },
]
</script>

<template>
  <div
    v-for="column of columns"
    :key="column.title"
  >
    <div class="title">
      {{ column.title }}
    </div>
    <div
      v-for="line of column.links"
      :key="line[0]"
      class="link-line"
    >
      <BcLink
        :to="line[2]"
        :target="line[3]"
        class="link"
      >
        <FontAwesomeIcon
          class="icon"
          :icon="line[1]"
        />
        {{ line[0] }}
      </BcLink>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";

.title {
  @include fonts.big_text_label;
  font-weight: var(--standard_text_medium_font_weight);
  color: var(--Light-Grey);
  line-height: 33px;

  @media (min-width: 600px) {
    margin-bottom: var(--padding);
  }

  @media (max-width: 599.9px) {
    margin-top: var(--padding);
  }
}

.link-line {
  @include fonts.standard_text;
  line-height: 27px;
}

.icon {
  display: inline-block;
  width: 20px;
  text-align: center;
}
</style>
