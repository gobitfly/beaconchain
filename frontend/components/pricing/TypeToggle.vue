<script lang="ts" setup>

</script>

<template>
  <div class="type-toggle-container">
    <div class="premium">
      <div class="text">
        {{ $t('pricing.premium') }}
      </div>
    </div>
    <div class="api-keys" disabled>
      <div class="text">
        {{ $t('pricing.API_keys') }}
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.type-toggle-container {
  width: 280px;
  height: 41px;
  margin-bottom: 37px;
  display: flex;
  align-items: center;
  user-select: none;

  // TODO: not interactive for beta launch
  .premium {
    flex: 1;
    height: 100%;
    display: flex;
    align-items: center;
    border-top-left-radius: 3px;
    border-bottom-left-radius: 3px;
    background: var(--button-color-active);

    .text {
      color: var(--primary-contrast-color);
      font-size: 12px;
      font-weight: var(--montserrat-semi-bold);
      padding-left: 18px;
    }
  }

  .api-keys {
    flex: 1;
    height: 100%;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--container-background);
    border-width: 1px 1px 1px 0;
    border-style: solid;
    border-color: var(--container-border-color);
    border-top-right-radius: 3px;
    border-bottom-right-radius: 3px;

    .text {
      color: var(--grey);
      font-size: 16px;
      font-weight: var(--montserrat-semi-bold);
    }
  }
}

@media (max-width: 1360px) {
  .type-toggle-container {
      .premium .text {
        font-size: 11px;
        font-weight: 700;
      }

      .api-keys .text {
        font-size: 11px;
        font-weight: 700;
      }
    }
}
</style>
