<script lang="ts" setup>
import { type PremiumPlan } from '~/types/pricing'

interface Props {
  plan?: PremiumPlan,
}
const props = defineProps<Props>()
</script>

<template>
  <div>
    [{{ props.plan?.Name }}]
  </div>
</template>

<style lang="scss" scoped>

</style>
