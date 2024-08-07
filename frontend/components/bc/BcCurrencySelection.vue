<script setup lang="ts">
import type { Currency } from '~/types/currencies'

defineProps<{
  showCurrencyIcon: boolean
}>()

const { currency, withLabel, setCurrency } = useCurrency()
</script>

<template>
  <BcDropdown
    v-model="currency"
    :options="withLabel"
    option-value="currency"
    option-label="label"
    variant="header"
    @update:model-value="(currency: Currency) => setCurrency(currency)"
  >
    <template #value>
      <span class="item in-header ">
        <span
          v-if="showCurrencyIcon"
          class="icon"
        >
          <IconCurrency
            v-if="currency"
            :currency="currency"
          />
        </span>{{ currency }}
      </span>
    </template>
    <template #option="slotProps">
      <span class="item">
        <span class="label">{{ slotProps.label }}</span>
        <span class="currency">{{ slotProps.currency }}</span>
        <span class="icon">
          <IconCurrency :currency="slotProps.currency" />
        </span>
      </span>
    </template>
  </BcDropdown>
</template>

<style lang="scss" scoped>
.item {
  display: flex;
  justify-content: space-between;
  gap: var(--padding);

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

  .currency {
    width: 30px;
    text-align: right;
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
