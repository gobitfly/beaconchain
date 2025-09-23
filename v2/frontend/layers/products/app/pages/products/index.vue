<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'

useHead({
  bodyAttrs: {
    // enforcing dark mode as light mode is not ready yet
    'data-theme': 'dark',
  },
})

const isOpen = ref(false)
const { t: $t } = useTranslation()
const id = {
  api: 'api',
  explorer: 'explorer',
  stakingHub: 'staking-hub',
}
const items: BaseNavigationItem[] = [
  {
    icon: 'file-code-2',
    label: $t('products.api'),
    to: `#${id.api}`,
  },
  {
    icon: 'coins',
    label: $t('products.staking_hub'),
    to: `#${id.stakingHub}`,
  },
  {
    icon: 'compass',
    label: $t('products.explorer'),
    to: `#${id.explorer}`,
  },
]
const navigationLeft = useTemplateRef<HTMLDialogElement>('navigationLeft')
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
      <section class="mt-11xl p-md">
        <h2 class="text-center">
          {{ $t('products.landing_page.search.title') }}
        </h2>
      </section>
      <section class="mt-11xl p-md">
        <h2
          :id="id.api"
          class="text-center "
        >
          {{ $t('products.landing_page.api.title') }}
        </h2>
      </section>
      <ProductLandingpageSection>
        <ProductLandingpageSectionStakinghub />
      </ProductLandingpageSection>
      <ProductLandingpageSection
        class="background-image"
      >
        <BaseHeading
          is="h2"
          :id="id.explorer"
          size="lg"
        >
          {{ $t('products.landing_page.explorer.title') }}
        </BaseHeading>
        <p>{{ $t('products.landing_page.explorer.description') }}</p>
        <BaseButton
          leading-icon="compass"
          trailing-icon="arrow-up-right"
          variant="branded"
          size="xl"
          to="/"
        >
          {{ $t('products.landing_page.explorer.action.go_to_explorer') }}
        </BaseButton>
        <div class="flex justify-center flex-wrap gap-6xl mt-7xl w-full">
          <span class="flex flex-col items-center min-w-[224px]">
            <BaseIcon
              name="clock"
              class="size-6xl"
            />
            <BaseText
              size="2xl"
              class="mt-4xl"
            >
              {{ $t('products.landing_page.explorer.transaction_tracking.title') }}
            </BaseText>
            <BaseText
              size="sm"
              class="mt-md"
              dimmed
            >
              {{ $t('products.landing_page.explorer.transaction_tracking.description') }}
            </BaseText>
          </span>
          <span class="flex flex-col items-center min-w-[224px]">
            <BaseIcon
              name="wallet"
              class="size-6xl"
            />
            <BaseText
              size="2xl"
              class="mt-4xl"
            >
              {{ $t('products.landing_page.explorer.wallet_tocken_insights.title') }}
            </BaseText>
            <BaseText
              size="sm"
              class="mt-md"
              dimmed
            >
              {{ $t('products.landing_page.explorer.wallet_tocken_insights.description') }}
            </BaseText>
          </span>
          <span class="flex flex-col items-center min-w-[224px]">
            <BaseIcon
              name="code"
              class="size-6xl"
            />
            <BaseText
              size="2xl"
              class="mt-4xl"
            >
              {{ $t('products.landing_page.explorer.smart_contract_analysis.title') }}
            </BaseText>
            <BaseText
              size="sm"
              class="mt-md"
              dimmed
            >
              {{ $t('products.landing_page.explorer.smart_contract_analysis.description') }}
            </BaseText>
          </span>
          <span class="flex flex-col items-center min-w-[224px]">
            <BaseIcon
              name="gas-station"
              class="size-6xl"
            />
            <BaseText
              size="2xl"
              class="mt-4xl"
            >
              {{ $t('products.landing_page.explorer.gas_fee_monitoring.title') }}
            </BaseText>
            <BaseText
              size="sm"
              class="mt-md"
              dimmed
            >
              {{ $t('products.landing_page.explorer.gas_fee_monitoring.description') }}
            </BaseText>
          </span>
        </div>
      </ProductLandingpageSection>
    </NuxtLayout>
  </div>
</template>

<style scoped>
.background-image {
  position: relative;

  &:before {
    z-index: -1;
    opacity: 0.1;
    content: '';
    position: absolute;
    inset: 0;
    background:
       linear-gradient(
      to bottom,
      var(--color-black) 0%,
      transparent 30%,
      transparent 70%,
      var(--color-black) 100%
    ),
    url('/assets-2usdf/img/bg-chain.webp');
    background-size: cover;
    background-position: center;
  }
}
</style>
