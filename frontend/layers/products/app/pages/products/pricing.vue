<script setup lang="ts">
import type { BaseNavigationItem } from '~/layers/base/app/components/BaseNavigationItem.vue'
import en from '#layers/products/i18n/locales/faq/en.json'

const { t: $t } = useTranslation()

const isOpen = ref(false)
const items: BaseNavigationItem[] = [
  {
    icon: 'file-code-2',
    label: $t('products.api.name'),
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

const messages = { 'en-US': en }
const { tm } = useI18n({
  // locale: 'en-US',
  messages,
})

const faqItems = tm('items') ?? []
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
      <ProductLandingpageSection class="mt-11xl">
        <BaseHeading
          is="h1"
          size="xl"
          class="text-center"
        >
          {{ $t('products.api.hero.title') }}
        </BaseHeading>
        <BaseText
          is="p"
          size="2xl"
          font-weight="bold"
          class="container text-center sm:text-balance"
        >
          {{ $t('products.api.hero.subtitle') }}
        </BaseText>
      </ProductLandingpageSection>
      <ProductLandingpageSection
        v-if="faqItems.length"
        class="mt-11xl"
      >
        <LazyProductPricingSectionFAQ
          :items="faqItems"
        />
      </ProductLandingpageSection>
    </NuxtLayout>
  </div>
</template>

<style scoped></style>
