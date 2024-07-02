<script setup lang="ts">
import type { ValidatorSubsetCategory } from '~/types/validator'

const count = ref(99)

const date = new Date()
date.setHours(date.getHours() + 5)

const list = computed(() => Array.from(Array(count.value)).map((_, index) => ({ index: index + 1, duty_objects: [Math.floor(date.getTime() / 1000), 230, 123] })))
const categories: ValidatorSubsetCategory[] = ['all', 'online', 'offline', 'pending', 'deposited', 'sync_current', 'sync_upcoming', 'sync_past', 'has_slashed', 'got_slashed', 'proposal_proposed', 'proposal_missed']

</script>

<template>
  <div>
    <div class="buttons">
      <Button @click="count=0">
        0
      </Button>
      <Button @click="count=1">
        1
      </Button>
      <Button @click="count=99">
        99
      </Button>
      <Button @click="count=100000">
        100000
      </Button>
    </div>

    <template v-for="category in categories" :key="category">
      <h3>{{ category }}</h3>
      <DashboardValidatorSubsetList :category="category" :validators="list" />
    </template>
  </div>
</template>
<style lang="scss" scoped>
.buttons{
  display: flex;
  gap: 10px;
  padding: 10px;
}
</style>
