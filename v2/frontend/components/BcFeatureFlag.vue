<script setup lang="ts">
import { warn } from 'vue'

type Feature = 'feature-account_dashboards'
type Environment = 'development' | 'production' | 'staging'

const currentEnvironment = useRuntimeConfig().public.deploymentType as Environment
if (!currentEnvironment) {
  warn('Environment variable `deploymentType` is not set.')
}

const staging: Feature[] = []
const activeFeatures: Record<Environment, Feature[]> = {
  development: [
    ...staging,
    'feature-account_dashboards',
  ],
  production: [],
  staging,
}

const props = defineProps<{
  feature: Feature,
}>()

const isEnabled = computed(
  () => activeFeatures[currentEnvironment]?.includes(props.feature),
)
</script>

<template>
  <slot v-if="isEnabled" />
</template>

<style scoped></style>
