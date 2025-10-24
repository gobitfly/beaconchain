<script setup lang="ts">
defineProps<{
  headline: string,
}>()
</script>

<template>
  <BaseHeading
    is="h2"
    size="lg"
    f
    class="text-center"
    :aria-label="`${headline} ${$t('base.common.ethereum')}`"
  >
    {{ headline }} <span class="text-link-500 before:mr-xs before:inline-block before:size-3xl before:bg-[url('/products/images/icon-eth.svg')] before:content-['']">ETH</span>
  </BaseHeading>
</template>
