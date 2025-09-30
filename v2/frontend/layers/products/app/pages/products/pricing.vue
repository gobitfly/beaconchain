<script setup lang="ts">
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'

useHead({
  bodyAttrs: {
    // enforcing dark mode as light mode is not ready yet
    'data-theme': 'dark',
  },
})
const { t: $t } = useTranslation()

const isOpen = ref(false)
const items: BaseNavigationItem[] = [
  {
    icon: 'file-code-2',
    label: $t('products.api'),
    to: '/product#api',
  },
  {
    icon: 'coins',
    label: $t('products.staking_hub'),
    to: '/products#staking-hub',
  },
  {
    icon: 'compass',
    label: $t('products.explorer'),
    to: 'products#explorer',
  },
]
</script>

<template>
  <div>
    <NuxtLayout name="base">
      <template #header>
        <BaseNavigation
          :items
          @open="isOpen = !isOpen"
        />
        <BaseNavigationLeft
          ref="navigationLeft"
          :items
          :is-open
          @close="isOpen = false"
        />
      </template>
      Page: /products/pricing
    </NuxtLayout>
  </div>
</template>

<style scoped></style>
