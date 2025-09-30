<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'
import { useBreakpoints } from '#layers/base/app/composables/useBreakpoints'

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
const { sm } = useBreakpoints()
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
      <div class="flex flex-col gap-11xl">
        <section class="mt-11xl p-md">
          <h2 class="text-center">
            {{ $t('products.landing_page.search.title') }}
          </h2>
        </section>
        <ProductLandingpageSection class="mt-11xl p-md">
          <BaseHeading
            is="h2"
            :id="id.api"
            size="lg"
          >
            {{ $t('products.landing_page.api.title') }}
          </BaseHeading>
          <p>{{ $t('products.landing_page.api.description') }}</p>
          <BaseButton
            leading-icon="file-code-2"
            trailing-icon="arrow-up-right"
            variant="branded"
            size="xl"
            to="/products/pricing"
          >
            {{ $t('products.landing_page.api.action.go_to_api') }}
          </BaseButton>
          <BaseCardLayout
            layout="2fr_1fr"
          >
            <BaseCard
              title-is="h3"
              :title="$t('products.landing_page.api.cards.api_docs.title')"
              :subtitle="$t('products.landing_page.api.cards.api_docs.subtitle')"
              title-icon="file-code-2"
              title-to="products/api/docs"
            >
              <div class="grid grid-cols-2 lg:grid-cols-3 gap-5xl mt-7xl">
                <article class="flex flex-col gap-4xl">
                  <ProductLandingIconApiExplorer
                    class="size-6xl"
                  />
                  <BaseHeading
                    is="h4"
                    size="xs"
                  >
                    {{ $t('products.landing_page.api.cards.api_docs.explorer_api.title') }}
                  </BaseHeading>
                  <BaseList
                    :items="[
                      $t('products.landing_page.api.cards.api_docs.explorer_api.list.real_time_transactions'),
                      $t('products.landing_page.api.cards.api_docs.explorer_api.list.historical_gas_prices'),
                      $t('products.landing_page.api.cards.api_docs.explorer_api.list.address_and_contract_analytics'),
                    ]"
                  />
                </article>
                <article class="flex flex-col gap-4xl">
                  <ProductLandingIconApiStakinghub
                    class="size-6xl"
                  />
                  <BaseHeading
                    is="h4"
                    size="xs"
                  >
                    {{ $t('products.landing_page.api.cards.api_docs.stakinghub_api.title') }}
                  </BaseHeading>
                  <BaseList
                    :items="[
                      $t('products.landing_page.api.cards.api_docs.stakinghub_api.list.validator_performance_metrics'),
                      $t('products.landing_page.api.cards.api_docs.stakinghub_api.list.real_time_staking_rewards_data'),
                      $t('products.landing_page.api.cards.api_docs.stakinghub_api.list.validator_status_and_lifecycle_events'),
                    ]"
                  />
                </article>
                <article class="flex flex-col gap-4xl">
                  <ProductLandingIconApiResourceMgmt
                    class="size-6xl"
                  />
                  <BaseHeading
                    is="h4"
                    size="xs"
                  >
                    {{ $t('products.landing_page.api.cards.api_docs.resource_management_api.title') }}
                  </BaseHeading>
                  <BaseList
                    :items="[
                      $t('products.landing_page.api.cards.api_docs.resource_management_api.list.add_remove_dashboards'),
                      $t('products.landing_page.api.cards.api_docs.resource_management_api.list.configure_validator_notifications'),
                    ]"
                  />
                </article>
              </div>
            </BaseCard>
            <BaseCard
              class="flex gap-3xl"
            >
              <BaseHeading
                is="h3"
                size="2xl"
                class="mt-auto"
              >
                {{ $t('products.landing_page.api.cards.api_key.title') }}
              </BaseHeading>
              <p>
                {{ $t('products.landing_page.api.cards.api_key.subtitle') }}
              </p>
              <template #footer>
                <BaseButton
                  size="xl"
                  variant="branded"
                  full
                  class="mt-7xl"
                  to="/products/api/key-management"
                >
                  {{ $t('products.landing_page.api.cards.api_key.action.get_api_key') }}
                </BaseButton>
              </template>
            </BaseCard>
            <BaseCard
              title-is="h3"
              :title="$t('products.landing_page.api.cards.api_pricing_plan.title')"
              title-icon="file-code-2"
              :subtitle="$t('products.landing_page.api.cards.api_pricing_plan.subtitle')"
            >
              <BaseButton
                v-if="sm"
                variant="quaternary"
                size="xl"
                trailing-icon="arrow-up-right"
                to="/products/pricing"
              >
                {{ $t('products.landing_page.api.cards.api_pricing_plan.action.compare_plans') }}
              </BaseButton>
              <template #footer>
                <BaseButton
                  v-if="!sm"
                  variant="secondary"
                  size="xl"
                  full
                  trailing-icon="arrow-up-right"
                  to="/products/pricing"
                >
                  {{ $t('products.landing_page.api.cards.api_pricing_plan.action.compare_plans') }}
                </BaseButton>
              </template>
            </BaseCard>
          </BaseCardLayout>
        </ProductLandingpageSection>
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
          <div class="flex justify-center flex-wrap gap-6xl mt-7xl w-full font-semibold">
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
                variant="secondary"
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
                variant="secondary"
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
                variant="secondary"
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
                variant="secondary"
              >
                {{ $t('products.landing_page.explorer.gas_fee_monitoring.description') }}
              </BaseText>
            </span>
          </div>
        </ProductLandingpageSection>
      </div>
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
