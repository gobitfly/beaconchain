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

      <section class="flex flex-col mt-11xl p-xl gap-4xl max-w-24xl items-center">
        <BaseCard
          class="text-center items-center p-5xl"
          gap="5xl"
        >
          <BaseHeading
            is="h2"
            size="lg"
          >
            {{ $t('products.api.plans.enterprise.title') }}
          </BaseHeading>
          <BaseText>
            {{ $t('products.api.plans.enterprise.description') }}
          </BaseText>
          <BaseButton
            leading-icon="file-code-2"
            trailing-icon="arrow-up-right"
            variant="branded"
            size="xl"
            to="/products/contact-sales"
          >
            {{ $t('base.common.contact_sales') }}
          </BaseButton>
          <BaseCardLayout
            layout="1fr_1fr_1fr"
            class="w-full"
            gap="5xl"
          >
            <BaseCard
              elevation="2"
              gap="xl"
              border-radius="xl"
            >
              <BaseIcon
                name="infinity"
                class="size-6xl mx-auto opacity-50"
              />
              <BaseHeading
                is="h3"
                size="md"
              >
                {{ $t('products.api.plans.enterprise.perks.rate_limit.title') }}
              </BaseHeading>
              <BaseText>{{ $t('products.api.plans.enterprise.perks.rate_limit.description') }}</BaseText>
            </BaseCard>
            <BaseCard
              elevation="2"
              gap="xl"
              border-radius="xl"
            >
              <BaseIcon
                name="headset"
                class="size-6xl mx-auto opacity-50"
              />
              <BaseHeading
                is="h3"
                size="md"
              >
                {{ $t('products.api.plans.enterprise.perks.dedicated_support.title') }}
              </BaseHeading>
              <BaseText>{{ $t('products.api.plans.enterprise.perks.dedicated_support.description') }}</BaseText>
            </BaseCard>
            <BaseCard
              elevation="2"
              gap="xl"
              border-radius="xl"
            >
              <BaseIcon
                name="shield"
                class="size-6xl mx-auto opacity-50"
              />
              <BaseHeading
                is="h3"
                size="md"
              >
                {{ $t('products.api.plans.enterprise.perks.sla.title') }}
              </BaseHeading>
              <BaseText>{{ $t('products.api.plans.enterprise.perks.sla.description') }}</BaseText>
            </BaseCard>
          </BaseCardLayout>
        </BaseCard>
      </section>
    </NuxtLayout>
  </div>
</template>

<style scoped></style>
