<script lang="ts" setup>
import { type ExtraDashboardValidatorsPremiumAddon } from '~/types/api/user'

interface Props {
  addon: ExtraDashboardValidatorsPremiumAddon,
  isYearly: boolean
}
defineProps<Props>()
</script>

<template>
  <div class="box-container">
    <div class="summary-container">
      <div class="validator-count">
        +1000 Validators
        <div class="subtext">
          per Dashboard i
        </div>
        <div class="per-validator">
          €0.0599 per Validator
        </div>
      </div>
    </div>
    <div class="price-container">
      <div class="price">
        <span>€59.99</span><span class="month">/month</span>
        <div class="year">
          719.88/year
        </div>
      </div>
      <div class="saving-info">
        You save €180 a year i
      </div>
      <div class="quantity-container">
        <div>
          Quanitity
        </div>
        <InputText class="input">
          1
        </InputText>
      </div>
      <Button label="Select Add-On" class="select-button" />
      <div class="footer">
        requires Orca plan
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.box-container {
  width: 348px;
  height: 100%;
  background-color: var(--container-background);
  border: 2px solid var(--container-border-color);
  border-radius: 7px;
  flex-shrink: 0;
  text-align: center;

  .summary-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    border-bottom: 2px solid var(--container-border-color);
    padding: 35px 0 26px 0;

    .validator-count {
      font-size: 24px;
      font-weight: 600;

      .subtext {
        font-weight: 400;
        margin-bottom: 16px;
      }

      .per-validator {
        color: var(--text-color-discreet);
        font-size: 20px;
        font-weight: 400;
      }
    }
  }

  .price-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 24px 34px 9px 34px;

    .price {
      font-size: 32px;
      font-weight: 600;
      margin-bottom: 28px;

      .month {
        color: var(--text-color-discreet);
        font-size: 17px;
        font-weight: 500;
      }

      .year {
        color: var(--text-color-discreet);
        font-size: 17px;
        font-weight: 500;
      }
    }

    .saving-info {
      width: 100%;
      display: flex;
      flex-direction: row;
      justify-content: center;
      align-items: center;
      gap: 13px;
      height: 37px;
      border-radius: 18px;
      background: var(--subcontainer-background);
      font-size: 17px;
      margin-bottom: 29px;
    }

    .quantity-container {
      display: flex;
      align-items: center;
      gap: 9px;
      font-size: 20px;
      margin-bottom: 32px;

      .input {
        width: 52px;
        border-radius: 9px;
      }
    }

    .select-button {
      width: 100%;
      height: 52px;
      font-size: 25px;
      font-weight: 500;
      margin-bottom: 26px;
    }

    .footer {
      width: 100%;
      text-align: right;
      font-size: 17px;
      font-weight: 400;
      color: var(--text-color-discreet);
    }
  }
}
</style>
