<script lang="ts" setup></script>

<template>
  <div class="header-line-container">
    <div class="header-line">
      <div class="title">
        {{ $t("pricing.premium") }}
      </div>
      <div class="subtitle">
        {{ $t("pricing.subtitle") }}
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.header-line-container {
  width: 100vw;
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  padding: 21px 0;
  border-top: 1px solid var(--container-border-color);
  border-bottom: 1px solid var(--container-border-color);
  background: var(--container-background);
  margin-bottom: 55px;

  .header-line {
    width: var(--pricing-content-width);
    display: flex;
    flex-direction: column;
    padding: 0 21px;

    .title {
      font-size: 18px;
      color: var(--primary-color);
      margin-bottom: 12px;
    }

    .subtitle {
      font-size: 32px;
    }
  }
}

@media (max-width: 1360px) {
  .header-line-container {
    padding-left: 5px;
    margin-bottom: 30px;

    .header-line .subtitle {
      font-size: 20px;
    }
  }
}
</style>
