<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'

const isOpen = ref(false)
const { t: $t } = useTranslation()
const id = {
  api: 'api',
  explorer: 'explorer',
  stackingHub: 'stacking-hub',
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
    to: `#${id.stackingHub}`,
  },
  {
    icon: 'compass',
    label: $t('products.explorer'),
    to: `#${id.explorer}`,
  },
]
const navigationLeft = useTemplateRef<HTMLDialogElement>('navigationLeft')
const { navigateToV1Login } = useV1Login()
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
      <section class="mt-11xl p-md">
        <h2
          :id="id.stackingHub"
          class="text-center "
        >
          {{ $t('products.landing_page.stacking_hub.title') }}
        </h2>
      </section>
      <section
        class="flex flex-col gap-4xl justify-center items-center mt-11xl text-center max-w-24xl mx-auto background-image p-md"
      >
        <BaseHeading
          is="h2"
          :id="id.explorer"
          size="lg"
          class=""
        >
          {{ $t('products.landing_page.explorer.title') }}
        </BaseHeading>
        <p>{{ $t('products.landing_page.explorer.description') }}</p>
        <BaseButton
          leading-icon="compass"
          trailing-icon="arrow-up-right"
          variant="branded"
          size="xl"
          @click="navigateToV1Login"
        >
          {{ $t('products.landing_page.explorer.action.go_to_explorer') }}
        </BaseButton>
        <div class="grid gap-6xl mt-7xl grid-cols-[repeat(auto-fit,minmax(224px,1fr))] w-full">
          <span class="flex flex-col items-center">
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
          <span class="flex flex-col items-center">
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
          <span class="flex flex-col items-center">
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
          <span class="flex flex-col items-center">
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
      </section>
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
