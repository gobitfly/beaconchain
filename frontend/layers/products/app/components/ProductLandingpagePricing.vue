<script setup lang="ts">
import type { Tab } from '~/layers/base/app/components/BaseTabs.vue'

const tabs: Tab[] = [
  {
    key: 'free',
    label: 'products.tiers.free',
  },
  {
    key: 'hobbyist',
    label: 'products.tiers.hobbyist',
  },
  {
    key: 'business',
    label: 'products.tiers.business',
  },
  {
    key: 'scale',
    label: 'products.tiers.scale',
  },
  {
    key: 'enterprise',
    label: 'products.tiers.enterprise',
  },
]

const { data } = useFetch('/api/bff/products/pricing')
const billingCycle = ref<'monthly' | 'yearly'>('monthly')
const { t: $t } = useTranslation()
const values = [
  {
    key: 'monthly',
    label: $t('products.landing_page.api.cards.api_pricing_plan.monthly'),
  },
  {
    key: 'yearly',
    label: $t('products.landing_page.api.cards.api_pricing_plan.yearly'),
  },
]
</script>

<template>
  <div class="flex flex-col gap-3xl justify-between">
    <section>
      <BaseSwitch
        v-model="billingCycle"
        screenreader-title="products.landing_page.api.cards.api_pricing_plan.billing_cycle"
        :values
        :class-list="{
          trackItem: 'py-md px-lg rounded-4xl text-gray-600 dark:text-gray-400  has-checked:text-white has-focus-visible:outline-2 has-focus-visible:outline-offset-2 has-focus-visible:outline-brand-300',
          track: 'p-px border border-transparent flex gap-md [&>*]:grow text-center text-md bg-white dark:bg-gray-950 rounded-4xl shadow-[0_1px_0.5px_0_rgba(255,255,255,0.08)_inset,0-1px_0_0_rgba(255,255,255,0.18)_inset] font-semibold',
          thumb: 'bg-gray-50 dark:bg-gray-800 rounded-4xl shadow-[0_2px_2px_0_rgba(0,0,0,0.25),_0_0.5px_0.5px_0_rgba(255,255,255,0.12)_inset]',
        }"
      />
    </section>
    <section>
      <BaseTabs
        :default-selected-tab="1"
        style="--avoid-layout-shift: 10rem;"
        class="min-h-(--avoid-layout-shift)"
        screenreader-title="products.landing_page.api.cards.api_pricing_plan.title"
        :class-list=" {
          tablist: 'p-xs font-semibold flex gap-xs bg-gray-100 dark:bg-black text-gray-600 dark:text-gray-400 rounded-4xl shadow-[0_-1px_0_0_rgba(255,255,255,0.18)_inset]',
          tab: 'p-sm grow rounded-4xl',
          activeTab: 'text-black dark:text-white',
          activeTabIndicator: 'dark:bg-gray-700 bg-gray-200 rounded-4xl text-white shadow-[0_2px_2px_0_rgba(0,0,0,0.25),_0_0.5px_0.5px_0_rgba(255,255,255,0.12)_inset]',
        }"
        :tabs
      >
        <template #tabpanel-free>
          <ProductLanindPagePricingDetails
            :variant="billingCycle"
            class="mt-xl"
            :price="0"
            :requests-per-second="data?.free.requests_per_second ?? 0"
          />
        </template>
        <template #tabpanel-hobbyist>
          <ProductLanindPagePricingDetails
            class="mt-xl"
            :price="(billingCycle === 'yearly' ? data?.hobbyist.yearly_price : data?.hobbyist.monthly_price) ?? 0"
            :price-with-discount="billingCycle === 'yearly' ? data?.hobbyist.yearly_price_with_discount : undefined"
            :variant="billingCycle"
            :requests-per-second="data?.hobbyist.requests_per_second ?? 0"
          />
        </template>
        <template #tabpanel-business>
          <ProductLanindPagePricingDetails
            class="mt-xl"
            :price="(billingCycle === 'yearly' ? data?.hobbyist.yearly_price : data?.hobbyist.monthly_price) ?? 0"
            :price-with-discount="billingCycle === 'yearly' ? data?.hobbyist.yearly_price_with_discount : undefined"
            :variant="billingCycle"
            :requests-per-second="data?.hobbyist.requests_per_second ?? 0"
          />
        </template>
        <template #tabpanel-scale>
          <ProductLanindPagePricingDetails
            class="mt-xl"
            :price="(billingCycle === 'yearly' ? data?.hobbyist.yearly_price : data?.hobbyist.monthly_price) ?? 0"
            :price-with-discount="billingCycle === 'yearly' ? data?.hobbyist.yearly_price_with_discount : undefined"
            :variant="billingCycle"
            :requests-per-second="data?.hobbyist.requests_per_second ?? 0"
          />
        </template>
        <template #tabpanel-enterprise>
          <div class="flex sm:grid-cols-[2fr_1fr] gap-lg mt-xl">
            <span
              class=""
            >{{ $t('products.landing_page.api.cards.api_pricing_plan.details.contact_sales_explainer') }}</span>
            <BaseButton
              class="h-fit self-center"
              size="lg"
              variant="branded"
              to="products/contact-sales"
            >
              {{ $t('products.landing_page.api.cards.api_pricing_plan.action.contact_sales') }}
            </BaseButton>
          </div>
        </template>
      </BaseTabs>
    </section>
  </div>
</template>

<style scoped></style>
