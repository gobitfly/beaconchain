<script setup lang="ts">
const { t: $t } = useI18n()
const enum Target {
  Internal = '_self',
  External = '_blank'
}

const columns = [
  {
    title: $t('footer.legal_notices'),
    links: [
      [$t('footer.imprint'), 'fas fa-building', '/imprint', Target.Internal],
      [$t('footer.terms'), 'fas fa-file-contract', 'https://storage.googleapis.com/legal.beaconcha.in/tos.pdf', Target.External],
      [$t('footer.privacy'), 'fas fa-user-secret', 'https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf', Target.External]
    ]
  },
  {
    title: $t('footer.resources'),
    links: [
      [$t('footer.api_docs'), 'fas fa-laptop-code', '/api/v2/docs/index.html', Target.Internal],
      [$t('footer.api_pricing'), 'fas fa-file-invoice-dollar', '/pricing', Target.Internal],
      [$t('footer.premium'), 'fas fa-user-astronaut', '/premium', Target.Internal],
      [$t('footer.advertise'), 'fas fa-ad', '/advertisewithus', Target.Internal],
      [$t('footer.shop'), 'fas fa-shopping-cart', 'https://shop.beaconcha.in', Target.External],
      [$t('footer.status'), 'fas fa-check-circle', 'https://status.beaconcha.in/', Target.External]
    ]
  },
  {
    title: $t('footer.links'),
    links: [
      ['Discord', 'fab fa-discord', 'https://dsc.gg/beaconchain', Target.External],
      ['Twitter', 'fab fa-twitter', 'https://twitter.com/beaconcha_in', Target.External],
      ['Github', 'fab fa-github', 'https://github.com/gobitfly/beaconchain', Target.External],
      ['Github Mobile App', 'fab fa-github', 'https://github.com/gobitfly/eth2-beaconchain-explorer-app', Target.External],
      [$t('footer.press_kit'), 'fas fa-newspaper', '/presskit', Target.Internal]
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
        <span class="icon"><span :class="line[1]" /></span>
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
  @media (min-width: 600px) { // large screen
    margin-bottom: 10px;
  }
  @media (max-width: 600px) { // mobile
    margin-top: 10px;
  }
}

.link-line {
  line-height: 27px;
}

.icon{
  display: inline-block;
  width: 20px;
  text-align: center;
}
</style>
