<script setup lang="ts">
const { t: $t } = useI18n()
const columns = [
  {
    title: $t('footer.legal_notices'),
    links: [
      [$t('footer.imprint'), ''],
      [$t('footer.terms'), ''],
      [$t('footer.privacy'), '']
    ]
  },
  {
    title: $t('footer.resources'),
    links: [
      [$t('footer.api_docs'), ''],
      [$t('footer.pricing'), ''],
      [$t('footer.advertise'), ''],
      ['Swag Shop', '']
    ]
  },
  {
    title: $t('footer.links'),
    links: [
      ['Discord', ''],
      ['Twitter', ''],
      ['Github', ''],
      ['Github Mobile App', ''],
      [$t('footer.press_kit'), '']
    ]
  }
]
</script>

<template>
  <div v-for="column in columns" :key="column.title">
    <div class="title">
      {{ column.title }}
    </div>
    <div v-for="link in column.links" :key="link[0]" class="link-line">
      <a :href="link[1]" class="link">{{ link[0] }}</a>
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
</style>
