<script lang="ts" setup>
const isYearly = ref(true)
</script>

<template>
  <BcPageWrapper>
    <div class="page-container">
      <div>
        Beaconcha.in/API Keys Toggle
      </div>
      <div class="header-line">
        <div class="premium">
          Beaconcha.in Premium
        </div>
        <h1 class="monitoring">
          Monitoring without limits on web and mobile.
        </h1>
      </div>
      <div>
        <BcToggle v-model="isYearly" true-option="Yearly" false-option="Monthly" />
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.page-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  .header-line {
    display: flex;
    flex-direction: column;
    padding: 21px 0 21px 360px;
    width: 100%;
    border: 1px solid var(--container-border-color);
    background: var(--container-background);

    .premium {
      @include fonts.subtitle_text;
      color: var(--primary-color);
    }

    .monitoring {
      margin-bottom: 0;
    }
  }
}
</style>
