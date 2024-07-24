<script lang="ts" setup>

</script>

<template>
  <div class="type-toggle-container">
    <div class="premium">
      <div class="label">
        {{ $t('pricing.premium') }}
      </div>
    </div>
    <div class="api-keys" disabled>
      <div class="label">
        {{ $t('pricing.API_keys') }}
      </div>
      <div class="subtext">
        {{ $t('common.coming_soon') }}
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.type-toggle-container {
  width: 280px;
  height: 50px;
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

    .label {
      color: var(--primary-contrast-color);
      font-size: 15px;
      font-weight: var(--montserrat-semi-bold);
      line-height: 1.1;
      padding-left: 18px;
    }
  }

  .api-keys {
    height: 100%;
    box-sizing: border-box;
    flex: 1;

    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;

    border-width: 1px 1px 1px 0;
    border-style: solid;
    border-color: var(--container-border-color);
    border-top-right-radius: 3px;
    border-bottom-right-radius: 3px;

    background: var(--container-background);
    color: var(--grey);
    font-weight: var(--montserrat-semi-bold);

    .label {
      padding-left: 18px;
      font-size: 16px;
    }

    .subtext {
      padding-left: 18px;
      font-size: 11px;
    }
  }
}

@media (max-width: 1360px) {
  .type-toggle-container {
    width: 240px;
    height: 43px;

    .api-keys .label,
    .premium .label {
      font-size: 11px;
      font-weight: var(--montserrat-bold);
    }

    .api-keys .subtext{
      font-size: 9px;
    }
  }
}
</style>
