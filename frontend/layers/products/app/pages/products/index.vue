<script setup lang="ts">
import type { BaseNavigationItem } from '#layers/base/app/components/BaseNavigationItem.vue'
import type { BlockchainSearchParams } from '~/layers/products/app/components/BlockchainSearchInput.vue'

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

const searchTypes: BlockchainSearchParams['types'] = [
  'address',
  'address_by_ens_name',
  'block',
  'ens_name',
  'epoch',
  'slot',
  'slot_by_block_root',
  'slot_by_state_root',
  'token',
  'transaction',
  'validators_by_deposit_address',
  'validators_by_graffiti',
  'validator_by_index',
  'validator_by_public_key',
  'validators_by_withdrawal_credential',
]
const searchParams = ref<BlockchainSearchParams>({
  input: '',
  networks: [
    1,
    560048,
  ],
  types: searchTypes,
})

const {
  data,
  error,
  execute,
  status,
} = useFetch('/api/bff/search', {
  body: searchParams,
  immediate: false,
  method: 'POST',
  watch: false,
})
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
          :items
          :is-open
          @close="isOpen = false"
        />
      </template>

      <ProductLandingpageSection class="px-[0px] sm:px-md gap-xl hero-section h-[70vh]">
        <div class="hero-section__background-effects" />
        <ProductLandingpageHeading
          :headline="$t('products.landing_page.search.title')"
        />
        <BlockchainSearchInput
          v-model="searchParams"
          class="w-[min(920px,100%)]"
          :results="data"
          :type-filters="searchTypes"
          :is-loading="status === 'pending'"
          :has-error="!!error"
          @search="execute()"
        />
      </ProductLandingpageSection>
      <ProductLandingpageSection class="mt-11xl">
        <ProductLandingpageHeading
          :id="id.api"
          :headline="$t('products.landing_page.api.title')"
        />
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
            title-to="/products/api/docs"
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
            class="sm:min-h-[20rem] sm:grid sm:grid-cols-[2fr_1fr] sm:grid-rows-[auto_auto_1fr]"
          >
            <ProductLandingpagePricing
              class="sm:col-start-2 sm:-col-end-1 sm:row-start-1 sm:-row-end-1"
            />
            <BaseButton
              variant="quaternary"
              size="xl"
              trailing-icon="arrow-up-right"
              to="/products/pricing"
              class="max-sm:hidden sm:size-fit sm:-row-end-1 self-end"
            >
              {{ $t('products.landing_page.api.cards.api_pricing_plan.action.compare_plans') }}
            </BaseButton>
            <template #footer>
              <BaseButton
                variant="secondary"
                size="xl"
                full
                trailing-icon="arrow-up-right"
                to="/products/pricing"
                class="sm:hidden col-span-2 row-span-2"
              >
                {{ $t('products.landing_page.api.cards.api_pricing_plan.action.compare_plans') }}
              </BaseButton>
            </template>
          </BaseCard>
        </BaseCardLayout>
      </ProductLandingpageSection>
      <ProductLandingpageSection class="mt-11xl">
        <ProductLandingpageSectionStakinghub />
      </ProductLandingpageSection>
      <ProductLandingpageSection class="mt-11xl explorer-section">
        <ProductLandingpageHeading
          :id="id.explorer"
          :headline="$t('products.landing_page.explorer.title')"
        />
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

<style scoped lang="scss">
.hero-section {
  position: relative;
}

.hero-section__background-effects {
  position: absolute;
  content: '';
  z-index: -2;
  inset: 0;
  margin: auto;
  background:
  linear-gradient(0deg, var(--color-black) 0%, rgba(16, 16, 16, 0.7) 85%, var(--color-black) 90%),
  url('/assets-2usdf/img/bg-hero.webp');
  background-size: cover;
  background-position: left top;

  &:before {
    position: absolute;
    content: '';
    z-index: -1;
    inset: 0;
    width: 100vw;
    backdrop-filter: blur(7px);
    transform: translateX(-50%);
    left: 50%;
  }
}

.explorer-section {
  position: relative;

  &:before {
    z-index: -1;
    opacity: 0.1;
    content: '';
    position: absolute;
    inset: 0;
    background:
      linear-gradient(to bottom,
        var(--color-black) 0%,
        transparent 30%,
        transparent 70%,
        var(--color-black) 100%),
      url('/assets-2usdf/img/bg-chain.webp');
    background-size: cover;
    background-position: center;
  }
}
</style>
