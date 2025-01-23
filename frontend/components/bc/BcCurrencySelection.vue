<script setup lang="ts">
const {
  exchangeRates,
  getCurrencyName,
  selectedCurrencyMain,
} = useCurrency()
const availableCurrencies = computed(
  () => exchangeRates.value.filter(({ code }) => code !== selectedCurrencyMain.value),
)
</script>

<template>
  <BcDropdown
    v-model="selectedCurrencyMain"
    :options="availableCurrencies"
    option-value="code"
    option-label="currency"
    variant="header"
    @update:model-value="(currencyCode: CurrencyCode) => selectedCurrencyMain = currencyCode"
  >
    <template #value>
      <span class="item in-header">
        <span
          class="icon"
        >
          <BcIconCurrency :currency-code="selectedCurrencyMain" />
        </span>
        {{ selectedCurrencyMain }}
      </span>
    </template>
    <template #option="{ code }">
      <span class="item">
        <span class="icon">
          <BcIconCurrency :currency-code="code" />
        </span>
        <span class="currency">{{ code }}</span>
        <span class="label">({{ getCurrencyName(code) }})</span>
      </span>
    </template>
  </BcDropdown>
</template>

<style lang="scss" scoped>
.item {
  display: flex;
  justify-content: space-between;
  gap: var(--padding-small);

  &.in-header {
    justify-content: flex-end;
    color: var(--light-grey);
    font-family: var(--main_header_font_family);
    font-size: var(--main_header_font_size);
    font-weight: var(--main_header_font_weight);
  }

  .label {
    flex-grow: 1;
  }

  .icon {
    height: 20px;
    width: 30px;
    display: flex;
    justify-content: flex-end;

    :deep(img),
    :deep(svg) {
      max-height: 100%;
      width: auto;
    }
  }

  &:not(.in-header) {
    .icon {
      justify-content: center;
    }
  }
}
</style>
