<script setup lang="ts">
const columns = [
  {
    title: 'Legal Notices',
    links: [
      ['Imprint', ''],
      ['Terms', ''],
      ['Privacy', '']
    ]
  },
  {
    title: 'Resources',
    links: [
      ['API Docs', ''],
      ['Pricing', ''],
      ['Advertise', ''],
      ['Swag Shop', '']
    ]
  },
  {
    title: 'Links',
    links: [
      ['Discord', ''],
      ['Twitter', ''],
      ['Github', ''],
      ['Github Mobile App', ''],
      ['Press Kit', '']
    ]
  }
]
</script>

<template>
  <div v-for="column in columns" :key="column.title">
    <div class="title">
      {{ column.title }}
    </div>
    <div v-for="link in column.links" :key="link[0]">
      <a :href="link[1]" class="link">{{ link[0] }}</a>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.title {
  color: var(--Light-Grey);
  font-size: 20px;
  font-weight: bold;
}
</style>
