<script setup lang="ts">
defineProps<{
  price: number,
  priceWithDiscount?: number,
  requestsPerSecond: number,
  variant: 'monthly' | 'yearly',
}>()
const { t: $t } = useTranslation()
</script>

<template>
  <div class="flex gap-lg">
    <section
      class="my-md"
    >
      <BaseText
        is="p"
        variant="secondary"
      >
        {{
          variant === 'yearly'
            ? $t('products.landing_page.api.cards.api_pricing_plan.details.cost_per_year')
            : $t('products.landing_page.api.cards.api_pricing_plan.details.cost_per_month') }}
      </BaseText>
      <div class="mt-lg flex items-baseline gap-lg font-bold">
        <template v-if="priceWithDiscount === undefined">
          <span class="text-2xl">
            {{ formatFiatCurrency(price, { trailingZeroDisplay: 'stripIfInteger' }) }}
          </span>
        </template>
        <template v-else>
          <BaseText
            is="span"
            v-if="priceWithDiscount"
            variant="secondary"
            aria-hidden="true"
            class="line-through"
          >
            {{ formatFiatCurrency(price, { trailingZeroDisplay: 'stripIfInteger' }) }}
          </BaseText>
          <span class="text-2xl">
            {{ formatFiatCurrency(priceWithDiscount, { trailingZeroDisplay: 'stripIfInteger' }) }}
          </span>
        </template>
      </div>
    </section>
    <section class="w-[1px] bg-gray-600 dark:bg-gray-200" />
    <section
      class="my-md"
    >
      <BaseText
        is="p"
        variant="secondary"
      >
        {{ $t('products.landing_page.api.cards.api_pricing_plan.details.requests_per_second') }}
      </BaseText>
      <div class="mt-lg text-2xl font-bold">
        {{ formatNumber(requestsPerSecond) }}
      </div>
    </section>
  </div>
</template>

<style scoped></style>
