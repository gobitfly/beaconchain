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
        <div class="monitoring">
          Monitoring without limits on web and mobile.
        </div>
      </div>
      <div class="toggle-container">
        <BcToggle v-model="isYearly" class="toggle" true-option="Yearly" false-option="Monthly" />
        <div class="save-up-text">
          SAVE UP TO 20%
        </div>
      </div>
      <div>
        [Guppy] [Dolphin] [Orca]
      </div>
    </div>
  </BcPageWrapper>
</template>

<style lang="scss" scoped>
@use '~/assets/css/fonts.scss';

.page-container {
  // The pricing page uses unique font styles
  font-family: var(--montserrat-family);
  font-weight: var(--montserrat-medium);

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
    margin-bottom: 55px;

    .premium {
      font-size: 18px;
      color: var(--primary-color);
      margin-bottom: var(--padding);
    }

    .monitoring {
      font-size: 32px;
    }
  }

  .toggle-container {
    display: flex;
    align-items: center;
    gap: var(--padding);
    margin-bottom: 55px;

    .toggle{
      font-size: 20px;
      margin-bottom: 0;
    }

    .save-up-text {
      width: 75px;
      color: var(--primary-color);
      text-align: center;
      font-size: 15px;
      font-weight: var(--montserrat-semi-bold);
    }
  }
}
</style>
