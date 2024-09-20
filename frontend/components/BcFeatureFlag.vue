<script setup lang="ts">
import type { FeatureFlag } from '~/types/feature-flags'

const props = defineProps<{
  feature: FeatureFlag,
}>()
const { has } = useFeatureFlag()
</script>

<template>
  <slot v-if="has(props.feature)" />
</template>

<style scoped></style>
