<script setup lang="ts">
const { t: $t } = useI18n()
const columns = [
  {
    title: $t('footer.legal_notices'),
    links: [
      [$t('footer.imprint'), '/imprint', 'fas fa-building'],
      [$t('footer.terms'), 'https://storage.googleapis.com/legal.beaconcha.in/tos.pdf', 'fas fa-file-contract'],
      [$t('footer.privacy'), 'https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf', 'fas fa-user-secret']
    ]
  },
  {
    title: $t('footer.resources'),
    links: [
      [$t('footer.api_docs'), '/api/v2/docs/index.html', 'fas fa-laptop-code'],
      [$t('footer.api_pricing'), '/pricing', 'fas fa-file-invoice-dollar'],
      [$t('footer.premium'), '/premium', 'fas fa-user-astronaut'],
      [$t('footer.advertise'), '/advertisewithus', 'fas fa-ad'],
      ['Swag Shop', 'https://shop.beaconcha.in', 'fas fa-shopping-cart'],
      [$t('footer.status'), 'https://status.beaconcha.in/', 'fas fa-check-circle']
    ]
  },
  {
    title: $t('footer.links'),
    links: [
      ['Discord', 'https://dsc.gg/beaconchain', 'fab fa-discord'],
      ['Twitter', 'https://twitter.com/beaconcha_in', 'fab fa-twitter'],
      ['Github', 'https://github.com/gobitfly/beaconchain', 'fab fa-github'],
      ['Github Mobile App', 'https://github.com/gobitfly/eth2-beaconchain-explorer-app', 'fab fa-github'],
      [$t('footer.press_kit'), '/presskit', 'fas fa-newspaper']
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
      <a :href="line[1]" class="link">
        <span class="icon"><span :class="line[2]" /></span>
        {{ line[0] }}
      </a>
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
