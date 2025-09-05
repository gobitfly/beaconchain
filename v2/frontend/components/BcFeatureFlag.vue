<script setup lang="ts">
const props = defineProps<{
  feature: FeatureFlag,
}>()
const { has } = useFeatureFlag()
</script>

<template>
  <slot v-if="has(props.feature)" />
</template>
