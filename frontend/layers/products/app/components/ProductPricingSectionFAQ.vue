<script setup lang="ts">
import type { AccordionItem } from '~/layers/base/app/components/BaseAccordion.vue'

const { items } = defineProps<{
  items: AccordionItem[],
}>()
</script>

<template>
  <div class="py-8xl flex w-full flex-col justify-center gap-4xl">
    <BaseHeading
      is="h2"
      id="staking-hub"
      size="lg"
      class="text-center"
    >
      {{ $t('products.pricing_page.faq.title') }}
    </BaseHeading>
    <BaseAccordion
      :items
    />
  </div>
</template>
