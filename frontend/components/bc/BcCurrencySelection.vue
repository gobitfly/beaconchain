<script setup lang="ts">
import type { Currency } from '~/types/currencies'

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
        <span class="icon">
          <IconCurrency v-if="currency" :currency="currency" />
        </span>{{ currency }}
      </span>
    </template>
    <template #option="slotProps">
      <span class="item">
        <span class="icon">
          <IconCurrency :currency="slotProps.currency" />
        </span>{{ slotProps.label }}
      </span>
    </template>
  </BcDropdown>
</template>
<style lang="scss" scoped>
.item {
  display: flex;
  justify-content: space-between;

  &.in-header {
    justify-content: flex-end;
    color: var(--light-grey);
    font-family: var(--main_header_font_size);
    font-size: var(--main_header_font_size);
    font-weight: var(--main_header_font_weight);
  }

  .icon {
    height: 20px;
    width: 30px;
    display: flex;
    justify-content: flex-end;
    margin-right: var(--padding);

    :deep(img),
    :deep(svg) {
      max-height: 100%;
      width: auto;
    }
  }
}
</style>
