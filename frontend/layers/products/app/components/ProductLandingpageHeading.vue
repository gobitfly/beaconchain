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
    {{ headline }} <span class="before:inline-block before:content-[''] before:mr-[0.25rem] before:bg-[url('/products/images/icon-eth.svg')] before:size-3xl text-link-500">ETH</span>
  </BaseHeading>
</template>
