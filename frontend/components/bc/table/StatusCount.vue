<script setup lang="ts">
import type { StatusCount } from '~/types/api/common'

interface Props {
  count?: StatusCount
}
defineProps<Props>()

</script>
<template>
  <span v-if="count">
    <span :class="{ positive: !!count.success }">{{ count.success ?? 0 }}</span>
    <span>/</span>
    <span :class="{ negative: !!count.failed }">{{ count.failed ?? 0 }}</span>
  </span>
</template>
